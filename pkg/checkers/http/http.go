package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/kdebug/pkg/base"
	"github.com/Azure/kdebug/pkg/env"
)

var (
	GoogleEndpoint = HttpEndpoint{
		Name: "Google HTTP endpoint",
		Url:  "https://google.com",
	}
	AzureIMDSEndpoint = HttpEndpoint{
		Name: "Azure IMDS HTTP endpoint",
		Url:  "http://169.254.169.254/metadata/versions",
	}
)

type HttpEndpoint struct {
	Name string
	Url  string
}

type HttpChecker struct {
}

func New() *HttpChecker {
	return &HttpChecker{}
}

func (c *HttpChecker) Name() string {
	return "Http"
}

func (c *HttpChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	results := []*base.CheckResult{}
	targets := getCheckTargets(ctx.Environment)
	var result *base.CheckResult
	for _, endpoint := range targets {
		res, err := http.Get(endpoint.Url)
		if err != nil {
			result = &base.CheckResult{
				Checker:     c.Name(),
				Error:       fmt.Sprintf("Fail to invoke HTTP GET method on URL %s.", endpoint.Url),
				Description: err.Error(),
				//todo: Recommendations and help links
			}
		} else {
			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode > 299 {
				res.Body.Close()
				var description string
				if err != nil {
					description = fmt.Sprintf("Failed to read response body: %s", err.Error())
				} else {
					description = fmt.Sprintf("Response body: %s", body)
				}
				result = &base.CheckResult{
					Checker:     c.Name(),
					Error:       fmt.Sprintf("HTTP GET on URL %s returns unsuccessful status %s .", endpoint.Url, res.Status),
					Description: description,
				}
			}else {
				result = &base.CheckResult{
					Checker:     c.Name(),
					Description: fmt.Sprintf("Successfully invoke HTTP GET on URL %s .", endpoint.Url),
				}
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func getCheckTargets(e env.Environment) []HttpEndpoint {
	targets := []HttpEndpoint{
		GoogleEndpoint,
	}

	if e.HasFlag("azure") {
		targets = append(targets, AzureIMDSEndpoint)
	}

	return targets
}
