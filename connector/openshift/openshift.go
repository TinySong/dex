package openshift

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/groups"
	"github.com/dexidp/dex/pkg/log"
	"github.com/dexidp/dex/storage/kubernetes/k8sapi"
)

const (
	wellKnownURLPath = "/.well-known/oauth-authorization-server"
	usersURLPath     = "/apis/user.openshift.io/v1/users/~"
)

// Config holds configuration options for OpenShift login
type Config struct {
	Issuer               string   `json:"issuer"`
	ClientID             string   `json:"clientID"`
	ClientSecret         string   `json:"clientSecret"`
	RedirectURI          string   `json:"redirectURI"`
	Groups               []string `json:"groups"`
	InsecureCA           bool     `json:"insecureCA"`
	RootCA               string   `json:"rootCA"`
	IncludeSystemRootCAs bool     `json:"includeSystemRootCAs"`
}

var (
	_ connector.CallbackConnector = (*openshiftConnector)(nil)
	_ connector.RefreshConnector  = (*openshiftConnector)(nil)
)

type openshiftConnector struct {
	apiURL               string
	redirectURI          string
	clientID             string
	clientSecret         string
	cancel               context.CancelFunc
	logger               log.Logger
	httpClient           *http.Client
	oauth2Config         *oauth2.Config
	insecureCA           bool
	rootCA               string
	includeSystemRootCAs bool
	groups               []string
}

type user struct {
	k8sapi.TypeMeta   `json:",inline"`
	k8sapi.ObjectMeta `json:"metadata,omitempty"`
	Identities        []string `json:"identities" protobuf:"bytes,3,rep,name=identities"`
	FullName          string   `json:"fullName,omitempty" protobuf:"bytes,2,opt,name=fullName"`
	Groups            []string `json:"groups" protobuf:"bytes,4,rep,name=groups"`
}

// Open returns a connector which can be used to login users through an upstream
// OpenShift OAuth2 provider.
func (c *Config) Open(id string, logger log.Logger) (conn connector.Connector, err error) {
	ctx, cancel := context.WithCancel(context.Background())

	wellKnownURL := strings.TrimSuffix(c.Issuer, "/") + wellKnownURLPath
	req, err := http.NewRequest(http.MethodGet, wellKnownURL, nil)

	openshiftConnector := openshiftConnector{
		apiURL:               c.Issuer,
		cancel:               cancel,
		clientID:             c.ClientID,
		clientSecret:         c.ClientSecret,
		insecureCA:           c.InsecureCA,
		logger:               logger,
		redirectURI:          c.RedirectURI,
		rootCA:               c.RootCA,
		includeSystemRootCAs: c.IncludeSystemRootCAs,
		groups:               c.Groups,
	}

	if openshiftConnector.httpClient, err = newHTTPClient(c.InsecureCA, c.RootCA, c.IncludeSystemRootCAs); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	var metadata struct {
		Auth  string `json:"authorization_endpoint"`
		Token string `json:"token_endpoint"`
	}

	resp, err := openshiftConnector.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to query OpenShift endpoint %v", err)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		cancel()
		return nil, fmt.Errorf("discovery through endpoint %s failed to decode body: %v",
			wellKnownURL, err)
	}

	openshiftConnector.oauth2Config = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL: metadata.Auth, TokenURL: metadata.Token,
		},
		Scopes:      []string{"user:info"},
		RedirectURL: c.RedirectURI,
	}
	return &openshiftConnector, nil
}

func (c *openshiftConnector) Close() error {
	c.cancel()
	return nil
}

// LoginURL returns the URL to redirect the user to login with.
func (c *openshiftConnector) LoginURL(scopes connector.Scopes, callbackURL, state string) (string, error) {
	if c.redirectURI != callbackURL {
		return "", fmt.Errorf("expected callback URL %q did not match the URL in the config %q", callbackURL, c.redirectURI)
	}
	return c.oauth2Config.AuthCodeURL(state), nil
}

type oauth2Error struct {
	error            string
	errorDescription string
}

func (e *oauth2Error) Error() string {
	if e.errorDescription == "" {
		return e.error
	}
	return e.error + ": " + e.errorDescription
}

// HandleCallback parses the request and returns the user's identity
func (c *openshiftConnector) HandleCallback(s connector.Scopes, r *http.Request) (identity connector.Identity, err error) {
	q := r.URL.Query()
	if errType := q.Get("error"); errType != "" {
		return identity, &oauth2Error{errType, q.Get("error_description")}
	}

	ctx := r.Context()
	if c.httpClient != nil {
		ctx = context.WithValue(r.Context(), oauth2.HTTPClient, c.httpClient)
	}

	token, err := c.oauth2Config.Exchange(ctx, q.Get("code"))
	if err != nil {
		return identity, fmt.Errorf("oidc: failed to get token: %v", err)
	}

	return c.identity(ctx, s, token)
}

func (c *openshiftConnector) Refresh(ctx context.Context, s connector.Scopes, oldID connector.Identity) (connector.Identity, error) {
	var token oauth2.Token
	err := json.Unmarshal(oldID.ConnectorData, &token)
	if err != nil {
		return connector.Identity{}, fmt.Errorf("parsing token: %w", err)
	}
	if c.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
	}
	return c.identity(ctx, s, &token)
}

func (c *openshiftConnector) identity(ctx context.Context, s connector.Scopes, token *oauth2.Token) (identity connector.Identity, err error) {
	client := c.oauth2Config.Client(ctx, token)
	user, err := c.user(ctx, client)
	if err != nil {
		return identity, fmt.Errorf("openshift: get user: %v", err)
	}

	if len(c.groups) > 0 {
		validGroups := validateAllowedGroups(user.Groups, c.groups)

		if !validGroups {
			return identity, fmt.Errorf("openshift: user %q is not in any of the required groups", user.Name)
		}
	}

	identity = connector.Identity{
		UserID:            user.UID,
		Username:          user.Name,
		PreferredUsername: user.Name,
		Email:             user.Name,
		Groups:            user.Groups,
	}

	if s.OfflineAccess {
		connData, err := json.Marshal(token)
		if err != nil {
			return identity, fmt.Errorf("marshal connector data: %v", err)
		}
		identity.ConnectorData = connData
	}

	return identity, nil
}

// user function returns the OpenShift user associated with the authenticated user
func (c *openshiftConnector) user(ctx context.Context, client *http.Client) (u user, err error) {
	url := c.apiURL + usersURLPath

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return u, fmt.Errorf("new req: %v", err)
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return u, fmt.Errorf("get URL %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return u, fmt.Errorf("read body: %v", err)
		}
		return u, fmt.Errorf("%s: %s", resp.Status, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return u, fmt.Errorf("JSON decode: %v", err)
	}

	return u, err
}

func validateAllowedGroups(userGroups, allowedGroups []string) bool {
	matchingGroups := groups.Filter(userGroups, allowedGroups)

	return len(matchingGroups) != 0
}

// newHTTPClient returns a new HTTP client
func newHTTPClient(insecureCA bool, rootCA string, includeSystemRootCAs bool) (*http.Client, error) {
	tlsConfig := tls.Config{}

	if insecureCA {
		tlsConfig = tls.Config{InsecureSkipVerify: true}
	} else if rootCA != "" {
		if !includeSystemRootCAs {
			tlsConfig = tls.Config{RootCAs: x509.NewCertPool()}
		} else {
			systemCAs, err := x509.SystemCertPool()
			if err != nil {
				return nil, fmt.Errorf("failed to read host CA: %w", err)
			}
			tlsConfig = tls.Config{RootCAs: systemCAs}
		}
		rootCABytes, err := os.ReadFile(rootCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read root-ca: %w", err)
		}
		if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCABytes) {
			return nil, fmt.Errorf("no certs found in root CA file %q", rootCA)
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}, nil
}
