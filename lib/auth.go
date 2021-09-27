package twitter

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

type Auth struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	HttpMethod        string
	BaseURL           string
	Query             string
}

func (a *Auth) FormatOAuth() string {
	nonce := getNonce()
	timestamp := getTimestamp()
	return fmt.Sprintf(
		`OAuth oauth_consumer_key="%s", oauth_nonce="%s", oauth_signature="%s", oauth_signature_method="%s", oauth_timestamp="%s", oauth_token="%s", oauth_version="%s"`,
		// `OAuth oauth_consumer_key="%s", oauth_consumer_secret="%s", oauth_token="%s", oauth_token_secret="%s", oauth_version="1.0"`,
		a.ConsumerKey,
		nonce,
		url.QueryEscape(a.getSig(nonce, timestamp)),
		"HMAC-SHA1",
		timestamp,
		a.AccessToken,
		"1.0",
	)
}

func getTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func (a *Auth) getSig(nonce string, timestamp string) string {
	sigkey := fmt.Sprintf("%s&%s", a.ConsumerSecret, a.AccessTokenSecret)
	parameterString := fmt.Sprintf(
		`oauth_consumer_key=%s&oauth_nonce=%s&oauth_signature_method=%s&oauth_timestamp=%s&oauth_token=%s&oauth_version=%s`,
		a.ConsumerKey,
		nonce,
		"HMAC-SHA1",
		timestamp,
		a.AccessToken,
		"1.0",
	)
	if a.Query != "" {
		parameterString = fmt.Sprintf(`%s&%s`, parameterString, a.Query)
	}
	fmt.Println("debug", parameterString)
	// for sort
	query, _ := url.ParseQuery(parameterString)
	parameterString = query.Encode()
	parameterString = strings.ReplaceAll(parameterString, "+", "%20")
	fmt.Println("parameterString", parameterString)
	message := fmt.Sprintf(
		`%s&%s&%s`,
		a.HttpMethod,
		url.QueryEscape(a.BaseURL),
		url.QueryEscape(parameterString),
	)
	fmt.Println("message: ", message)
	mac := hmac.New(sha1.New, []byte(sigkey))
	mac.Write([]byte(message))
	signedMac := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signedMac)
}

func getNonce() string {
	b := make([]byte, 32)
	rand.Seed(time.Now().UnixNano())
	_, _ = rand.Read(b)
	return url.QueryEscape(base64.StdEncoding.EncodeToString(b))
}
