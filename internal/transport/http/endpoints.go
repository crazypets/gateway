package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func makeHTTPProxyEndpoint(targetHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.ReverseProxy{
			Director: func(request *http.Request) {
				request.Header.Add("X-Forwarded-Host", request.Host)
				request.Header.Add("X-Origin-Host", targetHost)
				request.URL.Scheme = "http"
				request.URL.Host = targetHost
			},
		}

		var clientSpan opentracing.Span
		tracer := opentracing.GlobalTracer()

		if parentSpan := opentracing.SpanFromContext(c.Request.Context()); parentSpan != nil {
			clientSpan = tracer.StartSpan(
				c.Request.RequestURI,
				opentracing.ChildOf(parentSpan.Context()),
			)
		} else {
			clientSpan = tracer.StartSpan(c.Request.RequestURI)
		}
		defer clientSpan.Finish()

		ext.SpanKindRPCClient.Set(clientSpan)
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), clientSpan))

		if span := opentracing.SpanFromContext(c.Request.Context()); span != nil {
			opentracing.GlobalTracer().Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request.Header),
			)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func makeAuthenticateProxyEndpoint(proxyURL string) endpoint.Endpoint {
	tgt, _ := url.Parse(proxyURL)

	return httptransport.NewClient(
		"POST",
		tgt,
		encodePostAddressRequest,
		decodeDeleteAddressResponse,
		httptransport.ClientBefore(kitopentracing.ContextToHTTP(opentracing.GlobalTracer(), log.NewNopLogger())),
	).Endpoint()
}

func encodePostAddressRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(isAccessAllowedRequest)

	req.URL.Path = "/auth/is-access-allowed"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.Header)

	return encodeRequest(ctx, req, request)
}

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)

	return nil
}

func decodeDeleteAddressResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response isAccessAllowedResponse
	err := json.NewDecoder(resp.Body).Decode(&response)

	return response, err
}

type isAccessAllowedRequest struct {
	Header   string `json:"-"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type isAccessAllowedResponse struct {
	Ok bool `json:"ok"`
}
