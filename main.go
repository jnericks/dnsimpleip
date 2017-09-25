package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type options struct {
	token, account, zone, record string
}

func parseOptions() (options, error) {
	token := flag.String("token", "", "The API v2 OAuth token")
	account := flag.String("account", "", "Replace with your account ID")
	zone := flag.String("zone", "", "The zone ID is the name of the zone (or domain)")
	record := flag.String("record", "", "Replace with the Record ID")

	flag.Parse()

	if *token == "" ||
		*account == "" ||
		*zone == "" ||
		*record == "" {
		return options{}, fmt.Errorf("not all options passed in")
	}

	return options{
		token:   *token,
		account: *account,
		zone:    *zone,
		record:  *record,
	}, nil
}

func handleError(err error) {
	if err == nil {
		return
	}

	log.Println(err)
	os.Exit(1)
}

func main() {
	opts, err := parseOptions()
	handleError(err)

	ip, err := getIP()
	handleError(err)

	err = updateRecord(opts, ip)
	handleError(err)

	log.Printf("record %s for zone %s updated to %s\n", opts.record, opts.zone, ip)
}

func updateRecord(opts options, ip string) error {
	url := fmt.Sprintf("https://api.dnsimple.com/v2/%s/zones/%s/records/%s", opts.account, opts.zone, opts.record)
	body := fmt.Sprintf(`{"content":"%s"}`, ip)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", opts.token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("[%s] %s ", res.Status, string(b))
	}

	return nil
}

func getIP() (string, error) {

	const url = "http://icanhazip.com/"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(b), " \n"), nil
}
