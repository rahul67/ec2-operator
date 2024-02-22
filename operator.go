package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

func operateInstance(w http.ResponseWriter, r *http.Request) {
	var client string
	var instance string
	var action string
	var dryrun string

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: some issue getting body - %s", err)
	}
	if r.Method == "POST" {
		client = gjson.Get(string(body), "alerts.0.labels.client").String()
		instance = gjson.Get(string(body), "alerts.0.labels.instance").String()
		action = gjson.Get(string(body), "alerts.0.labels.action").String()
		dryrun = gjson.Get(string(body), "alerts.0.labels.dryrun").String()
	} else if r.Method == "GET" {
		client = r.URL.Query().Get("client")
		instance = r.URL.Query().Get("instance")
		action = r.URL.Query().Get("action")
		dryrun = r.URL.Query().Get("dryrun")
	}
	log.Printf("INFO: Request from %s - client: %s, action: %s, instance: %s, dryrun: %s\n", r.RemoteAddr, client, action, instance, dryrun)
	ec2ClientWrapper(client, action, instance, dryrun)
}

func operateHostname(w http.ResponseWriter, r *http.Request) {
	var client string
	var hostname string
	var action string
	var dryrun string

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: some issue getting body - %s", err)
	}
	if r.Method == "POST" {
		client = gjson.Get(string(body), "alerts.0.labels.client").String()
		hostname = gjson.Get(string(body), "alerts.0.labels.hostname").String()
		action = gjson.Get(string(body), "alerts.0.labels.action").String()
		dryrun = gjson.Get(string(body), "alerts.0.labels.dryrun").String()
	} else if r.Method == "GET" {
		client = r.URL.Query().Get("client")
		hostname = r.URL.Query().Get("hostname")
		action = r.URL.Query().Get("action")
		dryrun = r.URL.Query().Get("dryrun")
	}
	log.Printf("INFO: Request from %s - client: %s, action: %s, hostname: %s, dryrun: %s\n", r.RemoteAddr, client, action, hostname, dryrun)
	instance := ec2ClientWrapper(client, "findByIp", hostname, dryrun)
	instance = strings.Replace(instance, "\n", "", -1)
	instance = strings.Trim(instance, "\"")
	log.Printf("INFO: Got the instance: %s", instance)
	if len(instance) > 10 { // usually the output may be 'null' if an instance isn't found.
		ec2ClientWrapper(client, action, instance, dryrun)
	}
}

func main() {
	var defaultPort string
	var defaultHost string
	defaultPort = os.Getenv("PORT")
	if defaultPort == "" {
		log.Printf("INFO: Unable to get PORT from env. Assuming value: 8080")
		defaultPort = "8080"
	}
	// helpful in case you want to run it only on localhost
	defaultHost = os.Getenv("HOST")
	if defaultHost == "" {
		log.Printf("INFO: Unable to get HOST from env. Assuming value: %v", defaultHost)
	}

	log.Printf("INFO: Starting ec2-operator server on %s...\n", defaultPort)

	// In case of POST request, these APIs expect JSON payload in the format of webhook config in alertmanager:
	// https://prometheus.io/docs/alerting/latest/configuration/#webhook_config
	http.HandleFunc("/operateInstance", operateInstance)
	http.HandleFunc("/operateHostname", operateHostname)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: Request from %s - %s", r.RemoteAddr, r.RequestURI)
		w.Write([]byte("pong"))
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%v", defaultHost, defaultPort), nil)
	if err != nil {
		log.Println(err)
	}
}
