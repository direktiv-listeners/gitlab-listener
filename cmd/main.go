package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ENV_DEBUG          = "DIREKTIV_GITLAB_DEBUG"
	ENV_GITLAB_TOKEN   = "DIREKTIV_GITLAB_SECRET"
	ENV_DIREKTIV_TOKEN = "DIREKTIV_GITLAB_TOKEN"
	ENV_ENDPOINT       = "DIREKTIV_GITLAB_ENDPOINT"
	ENV_INSECURE       = "DIREKTIV_GITLAB_INSECURE_TLS"
	ENV_PATH           = "DIREKTIV_GITLAB_PATH"

	HEADER_GITLAB_TOKEN    = "X-Gitlab-Token"
	HEADER_GITLAB_EVENT    = "X-Gitlab-Event"
	HEADER_GITLAB_UUID     = "X-Gitlab-Event-UUID"
	HEADER_GITLAB_INSTANCE = "X-Gitlab-Instance"
)

var (
	localGitlabToken, endpoint string
)

func startServer() error {

	gin.SetMode(gin.ReleaseMode)

	// set logging
	debug := os.Getenv(ENV_DEBUG)
	if debug != "" {
		gin.SetMode(gin.DebugMode)
	}

	endpoint = os.Getenv(ENV_ENDPOINT)
	if os.Getenv("K_SINK") != "" {
		endpoint = os.Getenv("K_SINK")
	}

	if endpoint == "" {
		log.Fatal("endpoint for receiver not set")
	}

	log.Printf("using endpoint %s", endpoint)

	localGitlabToken = os.Getenv(ENV_GITLAB_TOKEN)

	path := os.Getenv(ENV_PATH)
	if path == "" {
		path = "/gitlab"
	}

	log.Printf("serving %s", path)

	r := gin.Default()
	r.POST(path, handleRequest)
	return r.Run()

}

func main() {

	log.Println("starting gitlab receiver")

	err := startServer()
	if err != nil {
		log.Fatalf("can not start server: %s", err.Error())
	}

}

func handleRequest(c *gin.Context) {

	gitlabToken := c.GetHeader(HEADER_GITLAB_TOKEN)

	// test token from gitlab
	if localGitlabToken != gitlabToken {
		log.Printf("gitlab token not valid")
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	payload := make(map[string]interface{})

	if err := c.BindJSON(&payload); err != nil {
		log.Printf("JSON invalid: %s", err.Error())
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	eventType := strings.ToLower(strings.ReplaceAll(c.GetHeader(HEADER_GITLAB_EVENT), " ", "-"))
	log.Printf("event type: %s", eventType)

	id, err := uuid.Parse(c.GetHeader(HEADER_GITLAB_UUID))
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		log.Printf("invalid event ID: %v.", err)
		return
	}
	log.Printf("event id: %s", id)

	instance := c.GetHeader(HEADER_GITLAB_INSTANCE)

	ce := cloudevents.NewEvent()

	pr, ok := payload["project"]
	if !ok {
		c.Writer.WriteHeader(http.StatusBadRequest)
		log.Printf("missing project field")
		return
	}
	project := pr.(map[string]interface{})
	source := fmt.Sprint(project["web_url"])

	ce.SetID(id.String())
	ce.SetType(eventType)
	ce.SetSource(source)
	ce.SetTime(time.Now())
	ce.SetDataContentType("application/json")
	ce.SetExtension("instance", instance)
	err = ce.SetData(payload)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		log.Printf("failed to set cloudevent data: %v", err)
		return
	}

	c.Writer.WriteHeader(200)
	go sendEvent(ce, endpoint)

}

func sendEvent(event cloudevents.Event, endpoint string) {

	skipTLS := false
	if os.Getenv(ENV_INSECURE) != "" {
		skipTLS = true
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS},
	}

	options := []cehttp.Option{
		cloudevents.WithTarget(endpoint),
		cloudevents.WithStructuredEncoding(),
		cloudevents.WithHTTPTransport(tr),
	}

	if len(os.Getenv(ENV_DIREKTIV_TOKEN)) > 0 {
		options = append(options,
			cehttp.WithHeader("Direktiv-Token", os.Getenv(ENV_DIREKTIV_TOKEN)))
	}

	t, err := cloudevents.NewHTTPTransport(
		options...,
	)
	if err != nil {
		log.Printf("unable to create transport: %s", err.Error())
		return
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Printf("unable to create client: %s", err.Error())
		return
	}

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Printf("unable to send cloudevent: %s", err.Error())
		return
	}

}
