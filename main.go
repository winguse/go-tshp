package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var proxyHost = flag.String("proxy-host", "example.com:443", "proxy url")
var username = flag.String("username", os.Getenv("SHP_USERNAME"), "proxy authentication username")
var password = flag.String("password", os.Getenv("SHP_PASSWORD"), "proxy authentication password")
var targetHost = flag.String("target-host", "destination-hostname:22", "host name and port number")

func main() {
	flag.Parse()

	pr, pw := io.Pipe()

	transport := http2.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	client := &http.Client{
		Transport: &transport,
	}

	request := http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Scheme: "https",
			Host:   *proxyHost,
		},
		Header: map[string][]string{
			"Proxy-Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(*username+":"+*password))},
		},
		Host: *targetHost,
		Body: pr,
	}

	response, err := client.Do(&request)
	if err != nil {
		log.Printf("error when sending request %s\n", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("Expected status OK, but %d\n", response.StatusCode)
		return
	}

	errCh := make(chan error, 2)

	defer pw.Close()
	defer os.Stdout.Close()
	defer os.Stdout.Close()
	defer response.Body.Close()

	go copy(pw, os.Stdin, errCh)
	go copy(os.Stdout, response.Body, errCh)

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil && err != io.EOF {
			log.Printf("Found transport error %s\n", err)
		}
	}
}

func copy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
