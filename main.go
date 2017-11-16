// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"strings"
)

type HTTP struct {
	Body       string            `json:"body"`
	Header     map[string]string `json:"header"`
	Method     string            `json:"method"`
	RemoteAddr string            `json:"remote_addr"`
	URL        string            `json:"url"`
}

type HTTPResponse struct {
	Body       string            `json:"body"`
	Header     map[string]string `json:"header"`
	StatusCode int               `json:"status_code"`
}

func main() {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("unable to load the event: %s", err)
	}

	var httpRequest HTTP

	err = json.Unmarshal(stdin, &httpRequest)
	if err != nil {
		log.Fatalf("unable to load the event: %s", err)
	}

	r := httptest.NewRequest(httpRequest.Method, httpRequest.URL, bytes.NewBufferString(httpRequest.Body))
	for k, v := range httpRequest.Header {
		r.Header.Add(k, v)
	}

	r.RemoteAddr = httpRequest.RemoteAddr

	w := httptest.NewRecorder()

	F(w, r)

	resp := w.Result()

	header := make(map[string]string)
	for k, v := range resp.Header {
		header[k] = strings.Join(v, ",")
	}

	out, err := json.Marshal(&HTTPResponse{
		Body:       w.Body.String(),
		Header:     header,
		StatusCode: resp.StatusCode,
	})

	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(out)
	os.Exit(0)
}
