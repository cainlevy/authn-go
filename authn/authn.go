package authn

import (
	"net/http"
	"time"
)

// TODO: jose/jwt references are all over the place. Refactor possible?

// Client provides JWT verification for ID tokens generated by the AuthN server. In the future it
// will also implement the server's private APIs (aka admin actions).
type Client struct {
	config   Config
	iclient  *internalClient
	verifier JWTClaimsExtractor
}

// NewClient returns an initialized and configured Client.
func NewClient(config Config) (*Client, error) {
	var err error
	config.setDefaults()

	ac := Client{}

	ac.config = config

	ac.iclient, err = newInternalClient(config.PrivateBaseURL, config.Username, config.Password)
	if err != nil {
		return nil, err
	}

	kchain := newKeychainCache(time.Duration(config.KeychainTTL)*time.Minute, ac.iclient)
	ac.verifier, err = NewIDTokenVerifier(config.Issuer, config.Audience, kchain)
	if err != nil {
		return nil, err
	}

	return &ac, nil
}

// SubjectFrom will return the subject inside the given idToken if and only if the token is a valid
// JWT that passes all verification requirements. The returned value is the AuthN server's account
// ID and should be used as a unique foreign key in your users data.
//
// If the JWT does not verify, the returned error will explain why. This is for debugging purposes.
func (ac *Client) SubjectFrom(idToken string) (string, error) {
	claims, err := ac.verifier.GetVerifiedClaims(idToken)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

//GetAccount gets the account with the associated id
func (ac *Client) GetAccount(id string) (*Account, error) { //Should this be a string or an int?
	return ac.iclient.GetAccount(id)
}

//Update updates the account with the associated id
func (ac *Client) Update(id, username string) error {
	return ac.iclient.Update(id, username)
}

//LockAccount locks the account with the associated id
func (ac *Client) LockAccount(id string) error {
	return ac.iclient.LockAccount(id)
}

//UnlockAccount unlocks the account with the associated id
func (ac *Client) UnlockAccount(id string) error {
	return ac.iclient.UnlockAccount(id)
}

//ArchiveAccount archives the account with the associated id
func (ac *Client) ArchiveAccount(id string) error {
	return ac.iclient.ArchiveAccount(id)
}

//ImportAccount imports an account with the provided information
func (ac *Client) ImportAccount(username, password string, locked bool) error {
	return ac.iclient.ImportAccount(username, password, locked)
}

//ExpirePassword expires the password of the account with the associated id
func (ac *Client) ExpirePassword(id string) error {
	return ac.iclient.ExpirePassword(id)
}

//ServiceStats gets the http response object from calling the service stats endpoint
func (ac *Client) ServiceStats() (*http.Response, error) {
	return ac.iclient.ServiceStats()
}

//ServerStats gets the http response object from calling the server stats endpoint
func (ac *Client) ServerStats() (*http.Response, error) {
	return ac.iclient.ServerStats()
}

// DefaultClient can be initialized by Configure and used by SubjectFrom.
var DefaultClient *Client

func defaultClient() *Client {
	if DefaultClient == nil {
		panic("Please initialize DefaultClient using Configure")
	}
	return DefaultClient
}

// Configure initializes the default AuthN client with the given config. This is necessary to
// use authn.SubjectFrom without keeping a reference to your own AuthN client.
func Configure(config Config) error {
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	DefaultClient = client
	return nil
}

// SubjectFrom will use the the client configured by Configure to extract a subject from the
// given idToken.
func SubjectFrom(idToken string) (string, error) {
	return defaultClient().SubjectFrom(idToken)
}
