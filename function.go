// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"cloud.google.com/go/compute/metadata"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sourcerepo/v1"
)

type ServiceAccountToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func F(w http.ResponseWriter, r *http.Request) {
	githubRepo, err := githubRepositoryFromRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	if err := mirrorGitHubCloudSourceRepositories(githubRepo); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
}

func mirrorGitHubCloudSourceRepositories(githubRepo *github.PushEventRepository) error {
	projectId := os.Getenv("GCP_PROJECT")
	if projectId == "" {
		return fmt.Errorf("The GCP_PROJECT environment variable must be set and non-empty")
	}

	serviceAccountEmail, serviceAccountToken, err := defaultServiceAccountCredentials()
	if err != nil {
		return err
	}

	ctx := context.Background()
	oauthHttpClient, err := google.DefaultClient(ctx, sourcerepo.SourceReadWriteScope)
	if err != nil {
		return fmt.Errorf("Unable to create source repo client: %s", err)
	}

	sourcerepoService, err := sourcerepo.New(oauthHttpClient)
	if err != nil {
		return fmt.Errorf("Unable to create source repo client: %s", err)
	}

	cloudSourceRepoName := fmt.Sprintf("projects/%s/repos/%s", projectId, *githubRepo.Name)

	_, err = sourcerepoService.Projects.Repos.Get(cloudSourceRepoName).Do()
	if err != nil {
		return fmt.Errorf("Unable to fetch the %s source repo : %s", cloudSourceRepoName, err)
	}

	tmpfile, err := ioutil.TempFile("", "git-credentials")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(fmt.Sprintf("https://%s:%s@source.developers.google.com",
		url.QueryEscape(serviceAccountEmail), serviceAccountToken))
	if err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "function")
	if err != nil {
		return fmt.Errorf("Unable to clone the %s repo: %s", *githubRepo.CloneURL, err)
	}

	githubRepoURL := *githubRepo.URL

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		parsedURL, err := url.Parse(githubRepoURL)
		if err != nil {
			return err
		}
		parsedURL.User = url.UserPassword(githubToken, "x-oauth-basic")
		githubRepoURL = parsedURL.String()
	}

	cmd := exec.Command("git", "clone", "--mirror", githubRepoURL, dir)
	output, err := cmd.CombinedOutput()
	log.Println(output)
	if err != nil {
		return fmt.Errorf("Unable to clone the %s repo: %s", *githubRepo.CloneURL, err)
	}

	cmd = exec.Command("git", "--git-dir", dir, "config", "credential.helper",
		fmt.Sprintf("store --file=%s", tmpfile.Name()))
	output, err = cmd.CombinedOutput()
	log.Println(output)
	if err != nil {
		return fmt.Errorf("Unable to create git credential.helper: %s", err)
	}

	remoteUrl := fmt.Sprintf("https://source.developers.google.com/p/%s/r/%s", projectId, *githubRepo.Name)

	cmd = exec.Command("git", "--git-dir", dir, "push", "--mirror", "--repo", remoteUrl)
	output, err = cmd.CombinedOutput()
	log.Println(output)
	if err != nil {
		return fmt.Errorf("Unable to sync the %s source repo: %s", cloudSourceRepoName, err)
	}

	return nil
}

func githubRepositoryFromRequest(r *http.Request) (*github.PushEventRepository, error) {
	payload, err := github.ValidatePayload(r, []byte("pipeline"))
	if err != nil {
		return nil, fmt.Errorf("Unable to validate webhook payload: %s", err)
	}

	webHookType := github.WebHookType(r)
	if webHookType != "push" {
		return nil, fmt.Errorf("The %s event type is not supported", webHookType)
	}

	event, err := github.ParseWebHook(webHookType, payload)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse webhook payload: %s", err)
	}

	return event.(*github.PushEvent).GetRepo(), nil
}

func defaultServiceAccountCredentials() (string, string, error) {
	email, err := metadata.Get("instance/service-accounts/default/email")
	if err != nil {
		return "", "", fmt.Errorf("Unable to get the default service account email: %s", err)
	}

	token, err := metadata.Get("instance/service-accounts/default/token")
	if err != nil {
		return "", "", fmt.Errorf("Unable to get the default service account token: %s", err)
	}

	var accessToken ServiceAccountToken
	if err := json.Unmarshal([]byte(token), &accessToken); err != nil {
		return "", "", fmt.Errorf("Unable to parse the default service account token: %s", err)
	}

	return email, accessToken.AccessToken, nil
}
