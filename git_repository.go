package gitmirror

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
	"path"
	"strings"
)

var (
	ErrGitRepositorySetURL = errors.New("string interpretation failed")
)

const (
	mirrorRemoteName = "mirror"
	gitAuthUser      = "git"
)

type IHub interface {
	SetAccount(account string)
	SetPassword(password string)
	Domain() string
	RepositoryExist(repository *GitRepository) bool
	RepositoryCreate(repository *GitRepository) error
	RepositoryFork(form *GitRepository, to *GitRepository) error
}

func GitAuthWithPublicKeysFromFile(pemFile, password string) transport.AuthMethod {
	auth, err := ssh.NewPublicKeysFromFile(gitAuthUser, pemFile, password)
	if err != nil {
		log.Fatalf("【GitMirror】Setting key failed %v", err)
		return nil
	}
	return auth
}
func GitAuthWithDefaultPublicKeysFromFile(password string) transport.AuthMethod {
	var err error
	homeDir, err := os.UserHomeDir()
	if err == nil {
		pemFile := path.Join(homeDir, ".ssh", "id_rsa")
		if FileExist(pemFile) {
			return GitAuthWithPublicKeysFromFile(pemFile, password)
		}
	}
	return nil
}
func GitAuthWithBasic(account, password string) transport.AuthMethod {
	return &http.BasicAuth{Username: account, Password: password}
}

type GitRepository struct {
	// IHub
	Hub IHub
	// pull push auth
	Auth transport.AuthMethod

	//true git@github.com:pkg6/gitmirror.git
	//false https://github.com/pkg6/gitmirror.git
	IsSSL bool
	//exp:
	//git@github.com:pkg6/git-mirror.git
	//https://github.com/pkg6/gitmirror.git
	RepositoryURL string
	//exp: pkg6
	OwnerOrOrg string
	//exp:git-mirror
	RepositoryName string
	//Edit repository details->Description
	Description string
	//Edit repository details->Website
	Homepage string
	//private | internal |public
	Visibility string
	//exp: pwd/git-mirror
	LocalPath string

	repository *git.Repository
}

func (r *GitRepository) SetHub(hub IHub) {
	r.Hub = hub
}

func (r *GitRepository) SetDefaultPublicKeys(password string) {
	r.SetAuth(GitAuthWithDefaultPublicKeysFromFile(password))
}

func (r *GitRepository) SetPublicKeys(pemFile, password string) {
	r.SetAuth(GitAuthWithPublicKeysFromFile(pemFile, password))
	return
}
func (r *GitRepository) SetBasicAuth(account, password string) {
	r.SetAuth(GitAuthWithBasic(account, password))
}

func (r *GitRepository) SetAuth(auth transport.AuthMethod) *GitRepository {
	r.Auth = auth
	return r
}

func (r *GitRepository) SetURL(repositoryURL string) error {
	var split []string
	r.RepositoryURL = repositoryURL
	if strings.Contains(repositoryURL, "git@") {
		URL := strings.Replace(strings.ReplaceAll(repositoryURL, "git@", ""), ":", "/", 1)
		split = strings.Split(URL, "/")
		r.IsSSL = true
	} else {
		index := strings.LastIndex(repositoryURL, "://")
		URL := repositoryURL[index+3:]
		split = strings.Split(URL, "/")
	}
	if l := len(split); l >= 3 {
		r.OwnerOrOrg = split[l-2]
		r.RepositoryName = strings.TrimSuffix(split[l-1], ".git")
		return nil
	}
	return ErrGitRepositorySetURL
}

func (r *GitRepository) URL() string {
	if r.RepositoryURL != "" {
		return r.RepositoryURL
	}
	if r.IsSSL {
		r.RepositoryURL = fmt.Sprintf("git@%s:%s/%s.git", r.Hub.Domain(), r.OwnerOrOrg, r.RepositoryName)
	} else {
		r.RepositoryURL = fmt.Sprintf("https://%s/%s/%s.git", r.Hub.Domain(), r.OwnerOrOrg, r.RepositoryName)
	}
	return r.RepositoryURL
}

func (r *GitRepository) GetLocalPath() string {
	if r.LocalPath != "" {
		return r.LocalPath
	}
	return r.RepositoryName
}

// MirrorClone
// git clone --mirror RepositoryURL LocalPath
func (r *GitRepository) MirrorClone() error {
	options := &git.CloneOptions{
		URL:    r.URL(),
		Mirror: true,
		Auth:   r.Auth,
	}
	_, err := git.PlainClone(r.GetLocalPath(), true, options)
	return err
}

// Open
// cd Name
func (r *GitRepository) Open() error {
	var err error
	r.repository, err = git.PlainOpen(r.GetLocalPath())
	return err
}

// Mirror
// git remote add mirror RepositoryURL
// git config --add remote.mirror.mirror true
func (r *GitRepository) Mirror() error {
	_, err := r.repository.CreateRemote(&config.RemoteConfig{
		Name:   mirrorRemoteName,
		URLs:   []string{r.URL()},
		Mirror: true,
	})
	return err
}

// MirrorPush
// git push --mirror
func (r *GitRepository) MirrorPush() error {
	return r.repository.Push(&git.PushOptions{
		RemoteName: mirrorRemoteName,
		Auth:       r.Auth,
		Force:      true,
		RefSpecs: []config.RefSpec{
			"+refs/heads/*:refs/heads/*",
			"+refs/tags/*:refs/tags/*",
		},
	})
}

func (r *GitRepository) Fork(dst *GitRepository) error {
	return r.Hub.RepositoryFork(r, dst)
}
