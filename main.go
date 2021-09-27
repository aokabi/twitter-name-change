package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	twitter "twitter-name-change/lib"

	"github.com/joho/godotenv"
)

var (
	client            = http.DefaultClient
	consumerKey       = ""
	consumerSecret    = ""
	accessToken       = ""
	accessTokenSecret = ""
	bearerToken       = ""
)

type ResponseRule struct {
	Data []twitter.Rule `json:"data"`
	Meta *twitter.Meta  `json:"meta"`
}

type ResponseTweet struct {
	Data *twitter.Tweet `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	consumerKey = os.Getenv("CONSUMER_KEY")
	consumerSecret = os.Getenv("CONSUMER_SECRET")
	accessToken = os.Getenv("ACCESS_TOKEN")
	accessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	bearerToken = os.Getenv("BEARER_TOKEN")
	// 一回ruleを全部削除する
	rules := getRules()
	deleteAllRules(rules)
	// searchのルールをセットする
	setRules()
	tmp := getRules()
	fmt.Println(tmp)
	// stream endpointに接続する
	err = connect()
	if err != nil {
		fmt.Printf("[error]: %w\n", err)
	}
	// readTimeline()
	// getSettings()
}

func getRules() []twitter.Rule {
	endpoint := "https://api.twitter.com/2/tweets/search/stream/rules"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Println(err)
	}
	resp, err := authorizationDo(req, 2)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	responseData := ResponseRule{
		make([]twitter.Rule, 0),
		&twitter.Meta{},
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Println(err)
	}
	return responseData.Data
}

func deleteAllRules(rules []twitter.Rule) {
	endpoint := "https://api.twitter.com/2/tweets/search/stream/rules"
	ids := make([]string, 0, len(rules))
	for _, v := range rules {
		ids = append(ids, v.ID)
	}
	type delete struct {
		IDs []string `json:"ids"`
	}
	payload := struct {
		Delete delete `json:"delete"`
	}{
		Delete: delete{ids},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}
	authorizationDo(req, 2)
}

func setRules() {
	endpoint := "https://api.twitter.com/2/tweets/search/stream/rules"
	rules := []twitter.Rule{
		{Value: "@aokabit"},
	}

	payload := struct {
		Add []twitter.Rule `json:"add"`
	}{
		rules,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := authorizationDo(req, 2)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	fmt.Printf("Status Code: %s\n", resp.Status)
	responseData := ResponseRule{
		make([]twitter.Rule, 0),
		&twitter.Meta{},
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Println(err)
		return
	}
}

func connect() error {
	endpoint := "https://api.twitter.com/2/tweets/search/stream"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("%w", err)
	}
	resp, err := authorizationDo(req, 2)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()
	fmt.Printf("Status Code: %s\n", resp.Status)
	fmt.Println("Header: ", resp.Header)
	for {
		responseData := ResponseTweet{
			&twitter.Tweet{},
		}
		r := bufio.NewReader(resp.Body)
		body, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("end of file")
			}
			log.Println(err)
			return fmt.Errorf("%w", err)
		}
		switch {
		case bytes.Equal(body, []byte{13, 10}): // heartbeat
			continue
		default:
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				log.Println(err)
				return fmt.Errorf("%w", err)
			}
		}

		fmt.Println("get response: ", responseData.Data)
		re := regexp.MustCompile("your name is ([^ ]+)")
		match := re.FindStringSubmatch(responseData.Data.Text)
		fmt.Printf("%q\n", match)
		if len(match) >= 1 {
			updateProfile(match[1])
			postTweet(fmt.Sprintf("My name is %s", match[1]))
		}
	}
}

func getSettings() {
	endpoint := "https://api.twitter.com/1.1/account/settings.json"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := authorizationDo(req, 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("getSettings: ", resp.Status)

}

func authorizeReq(req *http.Request, version int) (*http.Request, error) {
	if version <= 0 || version >= 3 {
		return nil, fmt.Errorf("invalid auth version")
	}
	switch version {
	case 1:
		auth := twitter.Auth{
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
			AccessToken:       accessToken,
			AccessTokenSecret: accessTokenSecret,
			HttpMethod:        req.Method,
			BaseURL:           fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path),
			Query:             req.URL.Query().Encode(),
		}
		req.Header.Add("authorization", auth.FormatOAuth())
	case 2:
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	}
	req.Header.Add("Content-type", "application/json")
	return req, nil
}

func authorizationDo(req *http.Request, version int) (*http.Response, error) {
	req, err := authorizeReq(req, version)
	if err != nil {
		return nil, err
	}
	fmt.Println(req)
	return client.Do(req)
}

func readTimeline() {
	endpoint := "https://api.twitter.com/1.1/statuses/home_timeline.json"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// query param
	q := req.URL.Query()
	q.Set("count", fmt.Sprintf("%d", 2))
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)

	// send request
	resp, err := authorizationDo(req, 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("updateProfile", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("body: ", body)
}

func updateProfile(name string) {
	endpoint := "https://api.twitter.com/1.1/account/update_profile.json"

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// query param
	q := req.URL.Query()
	q.Set("name", name)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)

	// send request
	resp, err := authorizationDo(req, 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("updateProfile", resp.Status)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
}

func postTweet(text string) {
	endpoint := "https://api.twitter.com/1.1/statuses/update.json"

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// query param
	q := req.URL.Query()
	q.Set("status", text)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)

	// send request
	resp, err := authorizationDo(req, 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

}
