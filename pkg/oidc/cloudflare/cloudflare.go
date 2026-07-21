package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/naiba/nezha/model"
	"github.com/naiba/nezha/pkg/utils"
	"github.com/naiba/nezha/service/singleton"
	"golang.org/x/oauth2"
)

type UserInfo struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

func OIDCBaseURL(endpoint, clientID string) string {
	return fmt.Sprintf("%s/cdn-cgi/access/sso/oidc/%s", strings.TrimRight(endpoint, "/"), clientID)
}

func OAuth2Config(endpoint, clientID, clientSecret, redirectURL string) *oauth2.Config {
	baseURL := OIDCBaseURL(endpoint, clientID)
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + "/authorization",
			TokenURL: baseURL + "/token",
		},
		RedirectURL: redirectURL,
	}
}

func FetchUserInfo(ctx context.Context, client *http.Client, endpoint, clientID string) (UserInfo, error) {
	var userInfo UserInfo
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OIDCBaseURL(endpoint, clientID)+"/userinfo", nil)
	if err != nil {
		return userInfo, fmt.Errorf("create Cloudflare userinfo request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return userInfo, fmt.Errorf("request Cloudflare userinfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return userInfo, fmt.Errorf("Cloudflare userinfo returned %s", resp.Status)
	}
	if err := utils.Json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return userInfo, fmt.Errorf("decode Cloudflare userinfo: %w", err)
	}
	if strings.TrimSpace(userInfo.Sub) == "" {
		return userInfo, fmt.Errorf("Cloudflare userinfo does not contain sub")
	}
	return userInfo, nil
}

func (u UserInfo) MapToNezhaUser() model.User {
	var user model.User
	login := strings.TrimSpace(u.Sub)
	singleton.DB.Where("login = ?", login).First(&user)
	user.Login = login
	user.Email = u.Email
	user.Name = u.Name
	return user
}
