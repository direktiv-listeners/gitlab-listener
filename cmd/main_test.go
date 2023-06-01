package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

var jsonData = []byte(`{
	"object_kind": "push",
	"event_name": "push",
	"before": "6827792928e1d42519c8815e1aaca1397b56a0e6",
	"after": "6827792928e1d42519c8815e1aaca1397b56a0e6",
	"ref": "refs/heads/main",
	"checkout_sha": "6827792928e1d42519c8815e1aaca1397b56a0e6",
	"message": null,
	"user_id": 1,
	"user_name": "Administrator",
	"user_username": "root",
	"user_email": null,
	"user_avatar": "https://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
	"project_id": 1,
	"project": {
	  "id": 1,
	  "name": "project",
	  "description": null,
	  "web_url": "http://264fe2e149ad/group/project",
	  "avatar_url": null,
	  "git_ssh_url": "git@264fe2e149ad:group/project.git",
	  "git_http_url": "http://264fe2e149ad/group/project.git",
	  "namespace": "group",
	  "visibility_level": 0,
	  "path_with_namespace": "group/project",
	  "default_branch": "main",
	  "ci_config_path": null,
	  "homepage": "http://264fe2e149ad/group/project",
	  "url": "git@264fe2e149ad:group/project.git",
	  "ssh_url": "git@264fe2e149ad:group/project.git",
	  "http_url": "http://264fe2e149ad/group/project.git"
	},
	"commits": [
	  {
		"id": "6827792928e1d42519c8815e1aaca1397b56a0e6",
		"message": "Initial commit",
		"title": "Initial commit",
		"timestamp": "2023-06-01T07:30:30+00:00",
		"url": "http://264fe2e149ad/group/project/-/commit/6827792928e1d42519c8815e1aaca1397b56a0e6",
		"author": {
		  "name": "Administrator",
		  "email": "[REDACTED]"
		},
		"added": [
		  "README.md"
		],
		"modified": [
  
		],
		"removed": [
  
		]
	  }
	],
	"total_commits_count": 1,
	"push_options": {
	},
	"repository": {
	  "name": "project",
	  "url": "git@264fe2e149ad:group/project.git",
	  "description": null,
	  "homepage": "http://264fe2e149ad/group/project",
	  "git_http_url": "http://264fe2e149ad/group/project.git",
	  "git_ssh_url": "git@264fe2e149ad:group/project.git",
	  "visibility_level": 0
	}
  }`)

var receiver testServer

func init() {

	receiver = testServer{}
	receiver.prepareReceiver()

	go receiver.startReceiver()

	os.Setenv("DIREKTIV_GITLAB_ENDPOINT", fmt.Sprintf("http://%s", receiver.addr))

	go startServer()

	time.Sleep(1 * time.Second)

}

func TestSending(t *testing.T) {

	request, err := http.NewRequest("POST", "http://127.0.0.1:8080/gitlab", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Gitlab-Event", "Push Hook")
	request.Header.Set("X-Gitlab-Instance", "http://264fe2e149ad")
	request.Header.Set("X-Gitlab-Event-UUID", "f46a4658-5f2b-4c87-bbfc-7bfa8acd271a")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer response.Body.Close()

	for {
		if receiver.lastRequest != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	t.Log(receiver.lastRequest)

}

type testServer struct {
	addr        string
	hasError    bool
	lastRequest map[string]interface{}
	lastHeaders map[string]string
}

func (s *testServer) prepareReceiver() {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("can not get listener")
	}
	defer l.Close()

	s.addr = l.Addr().String()

}

func (s *testServer) startReceiver() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		s.lastHeaders = make(map[string]string)

		for name, values := range r.Header {
			for _, value := range values {
				s.lastHeaders[name] = value
			}
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			s.hasError = true
			return
		}

		var resp map[string]interface{}

		err = json.Unmarshal(b, &resp)
		if err != nil {
			s.hasError = true
			return
		}

		s.lastRequest = resp

	})

	log.Fatal(http.ListenAndServe(s.addr, nil))

}
