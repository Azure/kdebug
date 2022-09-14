package aadssh

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	msal "github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	log "github.com/sirupsen/logrus"
)

// acquireTokenByAzureCLI acquires a token from AAD using Azure CLI credentials
func acquireTokenByAzureCLI(ctx context.Context, scopes []string, data map[string]string) (msal.AuthResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to get OS home dir: %+v", err)
	}

	tokenCacheFilePath := path.Join(homeDir, AzureCLIDirName, AzureCLITokenCacheFileName)
	f, err := os.Open(tokenCacheFilePath)
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to read Azure CLI token cache: %+v", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	var tokenCache struct {
		RefreshToken map[string]struct {
			CredentialType string `json:"credential_type"`
			Secret         string `json:"secret"`
			ClientID       string `json:"client_id"`
			HomeAccountID  string `json:"home_account_id"`
			Environment    string `json:"environment"`
		} `json:"RefreshToken"`
	}
	err = decoder.Decode(&tokenCache)
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to decode Azure CLI token cache: %+v", err)
	}

	var refreshToken string
	var tenantId string
	var clientId string
	var host string
	for _, token := range tokenCache.RefreshToken {
		// TODO: Add more checks
		if token.CredentialType == "RefreshToken" {
			refreshToken = token.Secret
			tenantId = strings.Split(token.HomeAccountID, ".")[1]
			host = token.Environment
			break
		}
	}

	if refreshToken == "" {
		return msal.AuthResult{}, fmt.Errorf("Cannot find any refresh token in Azure CLI token cache. Please do `az login`")
	}

	defaultScopes := []string{
		"openid",
		"profile",
		"offline_access",
	}
	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("grant_type", "refresh_token")
	values.Add("scope", strings.Join(append(scopes, defaultScopes...), " "))
	values.Add("refresh_token", refreshToken)
	for k, v := range data {
		values.Add(k, v)
	}
	bodyString := values.Encode()
	bodyStream := strings.NewReader(bodyString)

	url := fmt.Sprintf("https://%s/%s%s", host, tenantId, TokenURLSuffix)
	log.WithFields(log.Fields{"body": bodyString, "url": url}).Debug("Token request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyStream)
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to construct request: %+v", err)
	}

	httpClient := &http.Client{
		Timeout: time.Minute,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to request token: %+v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respContent, _ := ioutil.ReadAll(resp.Body)
		return msal.AuthResult{},
			fmt.Errorf("Unexpected token response status code: %d. Body: %s",
				resp.StatusCode, string(respContent))
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&body)
	if err != nil {
		return msal.AuthResult{}, fmt.Errorf("Fail to decode token response: %+v", err)
	}

	log.WithFields(log.Fields{"body": fmt.Sprintf("%+v", body)}).Debug("Token response")

	return msal.AuthResult{
		AccessToken: body.AccessToken,
		ExpiresOn:   time.Now().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, nil
}
