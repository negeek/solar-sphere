// Package gateway reverse-proxies incoming requests to the right backend
// service and normalizes the response. This is intentionally hand-rolled
// rather than built on net/http/httputil.ReverseProxy.
package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
)

// Gateway routes requests by their first path segment ("auth", "sentinel",
// ...) to the matching backend base URL.
type Gateway struct {
	serviceBaseURLs map[string]string
}

func NewGateway(serviceBaseURLs map[string]string) *Gateway {
	return &Gateway{serviceBaseURLs: serviceBaseURLs}
}

func (g *Gateway) Handle(w http.ResponseWriter, r *http.Request) {
	originalURL := r.URL.RequestURI()

	service := strings.Split(originalURL, "/")[1]
	baseURL, ok := g.serviceBaseURLs[service]
	if !ok {
		httpapi.JsonResponse(w, false, http.StatusBadRequest, "No such service to process request", nil)
		return
	}

	upstream := &upstreamRequest{
		Header: r.Header,
		Method: r.Method,
		URL:    fmt.Sprintf("%s%s", baseURL, originalURL),
		Body:   r.Body,
	}

	resp, err := doUpstreamRequest(upstream)
	if err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadGateway, err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	switch resp.Header.Get("Content-Type") {
	case "application/json":
		proxyJSON(w, resp)
	case "text/csv":
		proxyCSV(w, resp)
	default:
		httpapi.JsonResponse(w, false, http.StatusBadGateway, "Unrecognised Content-Type", nil)
	}
}

func proxyJSON(w http.ResponseWriter, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadGateway, err.Error(), nil)
		return
	}

	var data httpapi.Response
	if err := json.Unmarshal(body, &data); err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadGateway, err.Error(), nil)
		return
	}

	httpapi.JsonResponse(w, data.Success, data.StatusCode, data.Message, data.Data)
}

func proxyCSV(w http.ResponseWriter, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadGateway, err.Error(), nil)
		return
	}

	for headerKey, headerValues := range resp.Header {
		for _, headerValue := range headerValues {
			w.Header().Add(headerKey, headerValue)
		}
	}
	w.Write(body)
}

type upstreamRequest struct {
	Header http.Header
	Method string
	URL    string
	Body   io.Reader
}

func doUpstreamRequest(u *upstreamRequest) (*http.Response, error) {
	req, err := http.NewRequest(u.Method, u.URL, u.Body)
	if err != nil {
		return nil, err
	}
	for headerKey, headerValues := range u.Header {
		for _, headerValue := range headerValues {
			req.Header.Add(headerKey, headerValue)
		}
	}
	return http.DefaultClient.Do(req)
}
