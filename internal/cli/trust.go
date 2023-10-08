package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"scrutineer/internal/model"
)

func (c Cli) Whoami() error {
	req, err := c.getRequest("auth/whoami")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("Unauthorized. Please login first.")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("status code not OK")
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(buf.String())
	return nil
}

func (c Cli) GetRealm() error {
	req, err := c.getRequest("auth/realm")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New("Unauthorized. Please login first.")
		}
		return errors.New("status code not OK")
	}

	userRealm := model.UserRealm{}
	err = json.NewDecoder(resp.Body).Decode(&userRealm)
	if err != nil {
		return errors.New("could not decode response")
	}

	if len(userRealm.DirectTrust) == 0 {
		_, _ = fmt.Println("No direct trust relationships.")
		_, _ = fmt.Println("You can create one with `scrutineer trust user [user_id]`.")
		return nil
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Println("Direct trust relationships:")

		_, _ = fmt.Fprintln(w, "ID\tUser ID\tStart\tEnd")
		for _, trust := range userRealm.DirectTrust {
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", trust.SetId, trust.UserId, trust.StartTrust.Format(time.RFC3339), trust.EndTrust.Format(time.RFC3339))
		}
		_ = w.Flush()
	}

	return nil
}

func (c Cli) CreateUserUserTrust(trustedUserId string, start, end time.Time) error {
	if trustedUserId == "" {
		return errors.New("please specify which user you want to trust")
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(model.CreateUserUserTrust{
		TrustedUserId: trustedUserId,
		TrustStart:    start,
		TrustEnd:      end,
	})
	if err != nil {
		return err
	}

	req, err := c.postRequest("auth/trust/user", buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	fmt.Println("Successfully created trust relationship.")
	return nil
}

func (c Cli) DeleteUserUserTrust(trustId int) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(model.DeleteUserUserTrust{
		SetId: trustId,
	})
	if err != nil {
		return err
	}

	req, err := c.deleteRequest("auth/trust/user", buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	fmt.Println("Revoked trust relationship.")
	return nil
}
