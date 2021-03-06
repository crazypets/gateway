package http

import (
	"gateway/internal/config"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

const maxFileSize = 1024 * 1024 * 10

func NewHTTPServer(cfg *config.Config) *http.Server {
	router := gin.Default()
	router.Use(limits.RequestSizeLimiter(maxFileSize))

	defaultConfig := cors.DefaultConfig()
	defaultConfig.AddAllowHeaders("Authorization", "Access-Control-Expose-Headers")
	defaultConfig.AllowAllOrigins = true
	defaultConfig.AllowMethods = []string{"PUT", "POST"}

	swag.Register(swag.Name, &swagDoc{baseURL: cfg.SwaggerBaseURL})

	router.Use(cors.New(defaultConfig))
	router.Use(ginhttp.Middleware(opentracing.GlobalTracer()))
	router.Use(AuthorizeMiddleware(kitopentracing.TraceClient(opentracing.GlobalTracer(), "authorization")(makeAuthenticateProxyEndpoint(cfg.PermissionAddr))))

	for _, service := range cfg.ProxyServices {
		for _, endpoint := range service.Endpoints {
			switch endpoint.Method {
			case "GET":
				router.GET(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "HEAD":
				router.HEAD(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "POST":
				router.POST(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "PUT":
				router.PUT(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "DELETE":
				router.DELETE(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "OPTIONS":
				router.OPTIONS(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			case "PATCH":
				router.PATCH(endpoint.URI, makeHTTPProxyEndpoint(service.Addr))
			}
		}

	}

	router.GET("/health-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
}

type swagDoc struct {
	doc     string
	baseURL string
}

func (s *swagDoc) ReadDoc() string {
	if s.doc == "" {
		data, err := ioutil.ReadFile("./api/swagger/gateway.swagger.json")
		if err != nil {
			log.Println(err.Error())
			return ""
		}

		s.doc = strings.Replace(string(data), "BASE_URL", s.baseURL, 1)
	}
	return s.doc
}
