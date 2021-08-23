// Package ldap implements strategies for authenticating using the ldap protocol for authorizating tenxcloud provider.
package tenxcloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/log"
	"github.com/go-ldap/ldap/v3"
)

type Session struct {
	User UserInfo `json:"loginUser"`
}

type UserInfo struct {
	User        string `json:"user"`
	ID          int    `json:"id"`
	Namespace   string `json:"namespace"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Token       string `json:"token"`
	Role        int    `json:"role"`
	AccountType string `json:"accountType"`
}

type Config struct {
	// baseURL tec login address
	BaseURL     string `json:"baseURL"`
	RedirectURI string `json:"redirectURL"`
	// clientID for authorization
	ClientID string `json:"clientID"`
	// no useless
	ClientSecret string `json:"clientSecret"`

	UseLoginAsID bool `json:"useLoginAsID"`
	// check session validate
	SessionCheckURL string
}

const (
	loginURI   = "/oidc/login"
	sessionURL = "/ui/spi/v2/sessions/query"
	noLogin    = "not-logged-in"
	// config through dex config
	baseAuth = "8e059c94-f760-4f85-8910-202027cf0ff5"
)

type tenxConnector struct {
	Config
	//TODO useless
	TlsConfig *tls.Config

	Session Session
	Logger  log.Logger
}

var (
	_ connector.TenxConnector    = (*tenxConnector)(nil)
	_ connector.RefreshConnector = (*tenxConnector)(nil)
)

func (c *Config) Open(id string, logger log.Logger) (connector.Connector, error) {
	return &tenxConnector{
		Config: *c,
	}, nil
}

type refreshData struct {
	Username string     `json:"username"`
	Entry    ldap.Entry `json:"entry"`
}

// OpenConnector is the same as Open but returns a type with all implemented connector interfaces.
// func (c *Config) OpenConnector(logger log.Logger) (interface {
// 	connector.Connector
// 	connector.TenxConnector
// 	connector.RefreshConnector
// }, error) {
// 	return c.openConnector(logger)
// }

// func (c *Config) openConnector(logger log.Logger) (*tenxConnector, error) {
// 	return &tenxConnector{}, nil
// }
func (t *tenxConnector) Prompt() string {
	return "username"
}

func (t *tenxConnector) Login(ctx context.Context, s connector.Scopes, cookie string) (ident connector.Identity, err error) {
	// test
	// return connector.Identity{
	// 	UserID:            "1",
	// 	Username:          "admin",
	// 	PreferredUsername: "",
	// 	Email:             "admin@tenxcloud.com",
	// 	EmailVerified:     true,
	// }, nil

	if err := t.UserSession(cookie); err != nil {
		return connector.Identity{}, err
	}
	// call tce api use clienid and clientsecret
	//TODO get cookies from browser
	//TODO call tce api getting userinformation and filling in  Identity
	return connector.Identity{
		UserID:            fmt.Sprintf("%d", t.Session.User.ID),
		Username:          t.Session.User.User,
		PreferredUsername: "",
		Email:             t.Session.User.Email,
		EmailVerified:     true,
		Phone:             t.Session.User.Phone,
	}, nil
}

func (t *tenxConnector) Refresh(ctx context.Context, s connector.Scopes, ident connector.Identity) (connector.Identity, error) {
	return ident, nil
}

//session:tenx:[用户id]:[session ID]   代表已登录用户
func (t *tenxConnector) CheckCookie(cookies []*http.Cookie) (sessionid string, login bool) {
	// sessionid = "1:fc89179c736c44eca0ca56920032ca41"
	// login = false
	// return
	if cookies == nil {
		return
	}
	for _, c := range cookies {
		if c == nil {
			continue
		}
		if c.Name == "tce" {
			if strings.HasPrefix(c.Value, noLogin) {
				return
			}
			sessionid = c.Value
			login = true
			return
		}
	}
	fmt.Println("tenxConnector CheckCookie")
	// return default false value
	return
}

func (t *tenxConnector) UserSession(cookie string) error {
	jsonData := struct {
		ID string `json:"id"`
	}{ID: cookie}
	body, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}
	uri, err := url.Parse(t.Config.BaseURL)
	if err != nil {
		return err
	}
	uri.Path = path.Join(uri.Path, sessionURL)
	req, err := http.NewRequest("POST", uri.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	client := &http.Client{}
	if uri.Scheme == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("SYSTEM_CALL_SIGNATURE", baseAuth)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// fmt.Println(string(data))
	err = json.Unmarshal(data, &t.Session)
	if err != nil {
		return err
	}
	return nil
}

func (t *tenxConnector) Redirect() string {
	return path.Join(t.Config.BaseURL, loginURI)
}
