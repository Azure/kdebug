package aadssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	TokenURLSuffix = "/oauth2/v2.0/token"
)

// A HTTP trasport for adding additional parameters in AAD token request
type Transport struct {
	// Additional parameter key-value pairs
	data map[string]string
}

// RoundTrip modifies AAD token request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.WithFields(log.Fields{"url": *req.URL}).Debug("MSAL request")

	if strings.HasSuffix(req.URL.Path, TokenURLSuffix) {
		bodyBuf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		defer req.Body.Close()

		log.WithFields(log.Fields{"body": string(bodyBuf)}).Debug("Original request body")

		values, err := url.ParseQuery(string(bodyBuf))
		if err != nil {
			return nil, err
		}

		for k, v := range t.data {
			values.Add(k, v)
		}

		bodyString := values.Encode()
		log.WithFields(log.Fields{"body": bodyString}).Debug("Modified request body")

		bodyStream := strings.NewReader(bodyString)
		req.ContentLength = bodyStream.Size()
		req.Header.Set("Content-Length", fmt.Sprintf("%d", bodyStream.Size()))
		req.Body = io.NopCloser(bodyStream)
	}

	return http.DefaultTransport.RoundTrip(req)
}
