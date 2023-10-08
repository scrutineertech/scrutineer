package cli

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"scrutineer.tech/scrutineer/internal/model"
)

type Cli struct {
	serverUrl *url.URL
	Version   string
}

func New(serverUrl, version string) (Cli, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return Cli{}, err
	}
	return Cli{
		serverUrl: u,
		Version:   version,
	}, nil
}

func (c Cli) SignMsg(message, control, payload *os.File) { // usually message = Stdin, control = Stderr, payload = Stdout
	bufferedMessage := &bytes.Buffer{}
	if _, err := io.Copy(bufferedMessage, message); err != nil {
		fmt.Printf("failed to read message from stdin: %s\n", err)
		return
	}

	_, _ = control.WriteString("[GNUPG:] BEGIN_SIGNING\n")

	signature, err := c.serverSideSigning(bufferedMessage.Bytes())
	if err != nil {
		_, _ = control.WriteString(fmt.Sprintf("failed to sign message: %s\n", err))
		return
	}

	certSum := sha1.Sum(signature)
	certHex := hex.EncodeToString(certSum[:])

	_, _ = control.WriteString(fmt.Sprintf("[GNUPG:] SIG_CREATED S 19 8 00 %v %s\n", time.Now().Unix(), certHex))
	_, _ = payload.Write(signature)
}

func (c Cli) Verify(message, signature, control, payload *os.File) { // control = Stdout, message = Stdin, payload = Stderr
	// read and hash commit message
	messageBuffer := &bytes.Buffer{}
	if _, err := io.Copy(messageBuffer, message); err != nil {
		_, _ = payload.WriteString("scrutineer: failed to read commit ⚠️\n")
		os.Exit(1)
	}

	// tell git we are processing the signature
	_, _ = control.WriteString("[GNUPG:] FILE_START 1 signature\n") // 1 means "verify"
	signatureBuffer := &bytes.Buffer{}
	if _, err := io.Copy(signatureBuffer, signature); err != nil {
		_, _ = payload.WriteString("scrutineer: failed to read signature ⚠️\n")
		os.Exit(1)
	}
	_, _ = control.WriteString("[GNUPG:] FILE_DONE\n")

	// Where the verifying magic happens ✨
	verificationResponse, err := c.serverSideVerify(messageBuffer.Bytes(), signatureBuffer.Bytes())
	if err != nil {
		_, _ = payload.WriteString(fmt.Sprintf("scrutineer: failed to verify signature: %s ⚠️\n", err))
		os.Exit(1)
	}

	if verificationResponse.Verified {
		_, _ = control.WriteString("[GNUPG:] GOODSIG fingerprint username\ngpg: good commit\n")
		_, _ = payload.WriteString("scrutineer: verified ✅\n")
		os.Exit(0)
	}

	_, _ = control.WriteString("[GNUPG:] BADSIG fingerprint username\ngpg: bad commit\n")
	_, _ = payload.WriteString("scrutineer: BAD SIGNATURE ⚠️\n")
	os.Exit(1)
}

func (c Cli) serverSideSigning(message []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(model.SigningRequest{Message: message})
	if err != nil {
		return nil, err
	}

	req, err := c.postRequest("auth/sign", buf)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.New("Unauthorized. Please log in with 'scrutineer login'")
		}
		return nil, errors.New("did not receive Status Code 200")
	}

	var signingResponse model.SigningResponse
	err = json.NewDecoder(resp.Body).Decode(&signingResponse)
	if err != nil {
		return nil, err
	}

	return signingResponse.Signature, nil
}

func (c Cli) serverSideVerify(message, signature []byte) (model.VerificationResponse, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(model.VerificationRequest{Message: message, Signature: signature})
	if err != nil {
		return model.VerificationResponse{}, err
	}

	req, err := c.postRequest("auth/verify", buf)
	if err != nil {
		return model.VerificationResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.VerificationResponse{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return model.VerificationResponse{}, errors.New("did not receive Status Code 200")
	}

	var verificationResponse model.VerificationResponse
	err = json.NewDecoder(resp.Body).Decode(&verificationResponse)
	if err != nil {
		return model.VerificationResponse{}, err
	}

	return verificationResponse, nil
}
