package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
)

var (
	GoogleTarget = HttpTarget{
		Name: "Google HTTP endpoint",
		URL:  "https://google.com",
	}
	AzureIMDSTarget = HttpTarget{
		Name: "Azure IMDS HTTP endpoint",
		URL:  "http://169.254.169.254/metadata/versions",
		Header: http.Header{
			"Metadata": {"true"},
		},
	}
)

type HttpTarget struct {
	Name   string
	URL    string
	Header http.Header
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
		request, err := http.NewRequest("GET", httpTarget.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("Fail to create request for target %s: %+v",
				httpTarget.Name, err)
		}
		request.Header = httpTarget.Header

		response, err := c.Client.Do(request)
		if err != nil {
			result = &base.CheckResult{
				Checker: c.Name(),
				Error: fmt.Sprintf("Fail to invoke HTTP GET method on URL %s.",
					httpTarget.URL),
				Description: err.Error(),
				//todo: Recommendations and help links
			}
		} else {
			defer response.Body.Close()
			result = &base.CheckResult{
				Checker: c.Name(),
				Description: fmt.Sprintf("Successfully invoke HTTP GET on URL %s , response status code is %s.",
					httpTarget.URL, response.Status),
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
