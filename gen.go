// +build ignore

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

func main() {
	if len(os.Args) != 2 || !strings.HasPrefix(os.Args[1], "https://") {
		log.Fatal("usage: go run gen.go <url>")
	}
	url := os.Args[1]

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("expected 200, got", resp.StatusCode)
	}
	defer resp.Body.Close()

	bundle, err := ioutil.ReadAll(resp.Body)

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(bundle) {
		log.Fatalf("can't parse cerficiates from %s", url)
	}

	fp, err := os.Create("certifi.go")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	tmpl.Execute(fp, struct {
		Timestamp time.Time
		URL       string
		Bundle    string
	}{
		Timestamp: time.Now(),
		URL:       url,
		Bundle:    string(bundle),
	})
}

var tmpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// {{ .Timestamp }}
// {{ .URL }}

package gocertifi

//go:generate go run gen.go "{{ .URL }}"

import "crypto/x509"

const pemcerts string = ` + "`" + `
{{ .Bundle }}
` + "`" + `

// CACerts builds an X.509 certificate pool containing the
// certificate bundle from {{ .URL }} fetch on {{ .Timestamp }}.
// Returns nil on error along with an appropriate error code.
func CACerts() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM([]byte(pemcerts))
	return pool, nil
}
`))
