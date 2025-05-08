// mock-server – simple “echo instance-id” service on :9090
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	imdsv2TokenURL = "http://169.254.169.254/latest/api/token"
	instanceIDURL  = "http://169.254.169.254/latest/meta-data/instance-id"
)

// detectInstanceID tries (1) IMDS-v2, (2) IMDS-v1, (3) INSTANCE_ID env var,
// then (4) the container/host name.
func detectInstanceID() string {
	client := http.Client{Timeout: 500 * time.Millisecond}

	// 1) IMDS-v2
	tokenReq, _ := http.NewRequest(http.MethodPut, imdsv2TokenURL, nil)
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "60")
	if tokenResp, err := client.Do(tokenReq); err == nil && tokenResp.StatusCode == 200 {
		defer tokenResp.Body.Close()
		if token, err := io.ReadAll(tokenResp.Body); err == nil {
			idReq, _ := http.NewRequest(http.MethodGet, instanceIDURL, nil)
			idReq.Header.Set("X-aws-ec2-metadata-token", string(token))
			if idResp, err := client.Do(idReq); err == nil && idResp.StatusCode == 200 {
				defer idResp.Body.Close()
				if b, err := io.ReadAll(idResp.Body); err == nil {
					return string(b)
				}
			}
		}
	}

	// 2) IMDS-v1
	if resp, err := client.Get(instanceIDURL); err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		if b, err := io.ReadAll(resp.Body); err == nil {
			return string(b)
		}
	}

	// 3) Environment override
	if id := os.Getenv("INSTANCE_ID"); id != "" {
		return id
	}

	// 4) Host / container name
	if hn, err := os.Hostname(); err == nil {
		return hn
	}

	return "unknown-instance"
}

func main() {
	id := detectInstanceID()

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, id)
	})

	const port = "9090"
	log.Printf("mock-server ready (id=%s) on :%s", id, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
