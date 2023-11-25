package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cli/oauth"
	"net/http"
	"os"
	"path"
	"scrutineer.tech/scrutineer/internal/model"
)

func (c Cli) GithubAuth() error {
	fmt.Println("If you log in, you agree to the Terms of Service (https://scrutineer.tech/Terms.html)")

	flow := &oauth.Flow{
		Host:     oauth.GitHubHost("https://github.com"),
		ClientID: "9d59865c95e8efb2c29d", // scrutineertech Scrutineer app
		Scopes:   []string{},             // empty = only public data
	}

	accessToken, err := flow.DeviceFlow()
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(model.LoginRequest{
		AuthMethod: "github",
		AuthToken:  accessToken.Token,
	})
	if err != nil {
		return err
	}

	req, err := c.postUnauthRequest("login", buf)
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
		return errors.New("login failed")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("server error. Please try again later")
	}

	loginResp := model.LoginResponse{}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return err
	}

	err = c.setAuthCache(loginResp)
	if err != nil {
		return err
	}

	fmt.Printf(`Success! You are now logged in at Scrutineer.

There is one last thing. You need to set git to use Scrutineer
for signing and verifying commits.

Run the following commands in your terminal:

git config --global gpg.format x509
git config --global gpg.x509.program /path/to/scrutineer/bin

If you want all commits to be signed by default, run:
git config --global commit.gpgsign true

Or if you want to sign commits on a per-commit basis, run:
git commit -S -m "My commit message"

And that's all 🎉

------------------------------
Your user handle is: %s
You can always retrieve it with:
'scrutineer whoami'
------------------------------
`, loginResp.Username)

	return nil
}

type AuthConfig struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (c Cli) AuthCacheExists() error {
	if _, err := os.Stat(getAuthPath()); err != nil {
		return errors.New("Missing login credentials. Please login first.")
	}
	return nil
}

func (c Cli) getAuthCache() (AuthConfig, error) {
	authFile, err := os.Open(getAuthPath())
	if err != nil {
		return AuthConfig{}, err
	}
	defer func(authFile *os.File) {
		err := authFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(authFile)

	authConfig := AuthConfig{}
	err = json.NewDecoder(authFile).Decode(&authConfig)
	return authConfig, err
}

func (c Cli) setAuthCache(response model.TokenHolder) error {
	// Create dir if not exists
	err := os.MkdirAll(path.Dir(getAuthPath()), 0700)
	if err != nil {
		return err
	}

	authConfigFile, err := os.Create(getAuthPath())
	if err != nil {
		return err
	}

	err = json.NewEncoder(authConfigFile).Encode(AuthConfig{
		Username: response.GetUsername(),
		Token:    response.GetToken(),
	})
	if err != nil {
		return err
	}

	return authConfigFile.Close()
}

func (c Cli) Logout() error {
	req, err := c.postRequest("auth/logout", nil)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	_ = os.Remove(getAuthPath())
	fmt.Println("You are logged out now.")
	return nil
}

func (c Cli) DeleteMe() error {
	auth, err := c.getAuthCache()
	if err != nil {
		return err
	}

	fmt.Printf("Are you sure you want to delete your account %s? Type 'delete' to confirm.\n", auth.Username)
	var input string
	_, err = fmt.Scanln(&input)
	if err != nil {
		return err
	}
	if input != "delete" {
		return errors.New("aborting")
	}

	req, err := c.deleteRequest("auth/me", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("Something went wrong. Please try again later.")
	}

	_ = os.Remove(getAuthPath())
	fmt.Println("Your account was deleted")
	return nil
}

func getAuthPath() string {
	homeDir := os.Getenv("HOME")
	return path.Join(homeDir, ".config/scrutineer/auth")
}
