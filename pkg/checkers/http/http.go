package http

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
)

var (
	GoogleTarget = HttpTarget{
		Name: "Google HTTP endpoint",
		Request: &http.Request{
			URL: &url.URL{
				Scheme: "https",
				Host:   "google.com",
			},
		},
	}
	AzureIMDSTarget = HttpTarget{
		Name: "Azure IMDS HTTP endpoint",
		Request: &http.Request{
			URL: &url.URL{
				Scheme: "https",
				Host:   "169.254.169.254",
				Path:   "metadata/versions",
			},
			Header: map[string][]string{
				"Metadata": {"true"},
			},
		},
	}
)

type HttpTarget struct {
	Name    string
	Request *http.Request
}

type HttpChecker struct {
	Client HttpClient
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func New() *HttpChecker {
	return &HttpChecker{
		Client: &http.Client{
			// Disable proxy. Azure IMDS don't support to be used behind proxy.
			Transport: &http.Transport{Proxy: nil},
			Timeout:   10 * time.Second,
		},
	}
}

func (c *HttpChecker) Name() string {
	return "Http"
}

func (c *HttpChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}
	targets := getCheckTargets(ctx.Environment)
	var result *base.CheckResult
	for _, httpTarget := range targets {
		response, err := c.Client.Do(httpTarget.Request)
		if err != nil {
			result = &base.CheckResult{
				Checker:     c.Name(),
				Error:       fmt.Sprintf("Fail to invoke HTTP GET method on URL %s.", httpTarget.Request.RequestURI),
				Description: err.Error(),
				//todo: Recommendations and help links
			}
		} else {
			defer response.Body.Close()
			result = &base.CheckResult{
				Checker:     c.Name(),
				Description: fmt.Sprintf("Successfully invoke HTTP GET on URL %s , response status code is %s.", httpTarget.Request.RequestURI, response.Status),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func getCheckTargets(e env.Environment) []HttpTarget {
	targets := []HttpTarget{
		GoogleTarget,
	}

	if e.HasFlag("azure") {
		targets = append(targets, AzureIMDSTarget)
	}

	return targets
}
