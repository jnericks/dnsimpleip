package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
)

const (
	jsonip   = "http://ipv4.jsonip.com/"
	dnsimple = "https://api.dnsimple.com/v1/"
)

type options struct {
	Email    string `long:"email"  description:"DNSimple Email"         required:"true"`
	ApiToken string `long:"token"  description:"DNSimple v1 Api Token"  required:"true"`
	Domain   string `long:"domain" description:"DNSimple Domain Name"   required:"true"`
	RecordID int    `long:"record" description:"DNSimple DNS Record ID" required:"true"`
}

func getIP() (string, error) {
	resp, err := http.Get(jsonip)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	var j struct {
		IP string `json:"ip"`
	}

	json.NewDecoder(resp.Body).Decode(&j)
	return j.IP, nil
}

/*
curl  -H 'X-DNSimple-Token: <email>:<token>' \
      -H 'Accept: application/json' \
      -H 'Content-Type: application/json' \
      -X PUT \
      -d '<json>' \
      https://api.dnsimple.com/v1/domains/example.com/records/2
*/
func updateRecord(o options, ip string) error {
	url := fmt.Sprintf("https://api.dnsimple.com/v1/domains/%s/records/%d", o.Domain, o.RecordID)
	json := fmt.Sprintf(`{"record":{"content":"%s"}}`, ip)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}

	req.Header.Set("X-DNSimple-Token", fmt.Sprintf("%s:%s", o.Email, o.ApiToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DNSimple update request responded with %s", resp.Status)
	}

	return nil
}

func main() {
	var (
		opts options
		ip   string
		err  error
	)

	handleError := func(err error) {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	_, err = flags.Parse(&opts)
	handleError(err)

	ip, err = getIP()
	handleError(err)

	err = updateRecord(opts, ip)
	handleError(err)

	fmt.Println("DNSimple record updated successfully")
}
