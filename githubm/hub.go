package githubm

import (
	"context"
	"github.com/google/go-github/v57/github"
	"github.com/pkg6/gitmirror"
	"net/http"
	"strings"
)

type Hub struct {
	Username, Password string
}

func (h *Hub) Domain() string {
	return "github.com"
}

func (h *Hub) SetAccount(account string) {
	h.Username = account
}
func (h *Hub) SetPassword(password string) {
	h.Password = password
}

func (h *Hub) Client() *github.Client {
	var httpClient *http.Client
	//https://github.com/settings/tokens
	if h.Username != "" && h.Password != "" {
		basicAuth := github.BasicAuthTransport{
			Username: strings.TrimSpace(h.Username),
			Password: strings.TrimSpace(h.Password),
		}
		httpClient = basicAuth.Client()
	}
	return github.NewClient(httpClient)
}

func (h *Hub) RepositoryExist(repository *gitmirror.GitRepository) bool {
	_, resp, err := h.Client().Repositories.Get(
		context.Background(),
		repository.OwnerOrOrg,
		repository.RepositoryName,
	)
	if resp.StatusCode != http.StatusOK || err != nil {
		return false
	}
	return true
}

func (h *Hub) RepositoryCreate(repository *gitmirror.GitRepository) error {
	client := h.Client()
	if repository.OwnerOrOrg == h.Username {
		repository.OwnerOrOrg = ""
	}
	gr := &github.Repository{
		Name: &repository.RepositoryName,
	}
	if repository.Description != "" {
		gr.Description = &repository.Description
	}
	if repository.Homepage != "" {
		gr.Homepage = &repository.Homepage
	}
	if repository.Visibility != "" {
		gr.Visibility = &repository.Visibility
	}
	_, _, err := client.Repositories.Create(context.Background(), repository.OwnerOrOrg, gr)
	return err
}
func (h *Hub) RepositoryDelete(repository *gitmirror.GitRepository) error {
	client := h.Client()
	_, err := client.Repositories.Delete(context.Background(), repository.OwnerOrOrg, repository.RepositoryName)
	return err
}

func (h *Hub) RepositoryFork(form *gitmirror.GitRepository, to *gitmirror.GitRepository) error {
	opts := &github.RepositoryCreateForkOptions{}
	if to != nil {
		opts.Organization = to.OwnerOrOrg
		opts.Name = to.RepositoryName
	}
	_, _, err := h.Client().Repositories.CreateFork(context.Background(),
		form.OwnerOrOrg,
		form.RepositoryName,
		opts,
	)
	return err
}
