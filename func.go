package gitmirror

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"log"
	"os"
	"path/filepath"
)

func MigrateRepository(form *GitRepository, to *GitRepository) error {
	if form.Hub != nil && to.Hub != nil {
		if form.Hub.Domain() == to.Hub.Domain() {
			return ForkRepository(form, to)
		}
	}
	return MirrorPushRepository(form, to)
}

// ForkRepository
// gr := &gitmirror.GitRepository{}
// gr.SetHub(&githubm.Hub{Username: "username", Password: "ghp_**************"})
// gr.SetURL("https://github.com/pkg6/gitmirror.git")
// gitmirror.ForkRepository(gr,nil)
func ForkRepository(form *GitRepository, to *GitRepository) error {
	log.Printf("【GitMirror】fork FORM : %s TO : %s", form.URL(), to.URL())
	return form.Fork(to)
}
func SimpleForkRepository(hub IHub, repositoryURL string) error {
	gr := &GitRepository{}
	gr.SetHub(hub)
	if err := gr.SetURL(repositoryURL); err != nil {
		return err
	}
	return ForkRepository(gr, nil)
}

// MirrorPushRepository
// form := &gitmirror.GitRepository{}
// form.SetURL("https://github.com/pkg6/gitmirror.git")
// h := &githubm.Hub{Username: "username", Password: "ghp_****************"}
// to := &gitmirror.GitRepository{}
// to.SetHub(h)
// to.SetBasicAuth("account", "ghp_****************") || to.SetDefaultPublicKeys("")
// to.SetURL("git@github.com:you/gitmirror.git")
// gitmirror.MirrorPushRepository(form, to)
func MirrorPushRepository(form *GitRepository, to *GitRepository) error {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, form.GetLocalPath())
	defer func() {
		_ = os.RemoveAll(path)
	}()
	log.Printf("【GitMirror】Starting git clone --mirror %s %s", form.URL(), form.GetLocalPath())
	if err := form.MirrorClone(); err != nil {
		return err
	}
	log.Printf("【GitMirror】Clone completed: %s", path)
	to.LocalPath = path
	err := to.Open()
	if err != nil {
		return err
	}
	err = to.Mirror()
	if err != nil {
		return err
	}
	log.Printf("【GitMirror】git remote add mirror %s", to.URL())
	log.Printf("【GitMirror】git config --add remote.mirror.mirror true")
	if to.Hub != nil {
		if !to.Hub.RepositoryExist(to) {
			log.Printf("【GitMirror】RepositoryURL %s does not exist ", to.URL())
			err = to.Hub.RepositoryCreate(to)
			if err != nil {
				return err
			}
			log.Printf("【GitMirror】Create RepositoryURL %s", to.URL())
		}
	}
	return to.MirrorPush()
}
func SimpleMirrorPushRepository(formRepositoryURL, toRepositoryURL string, toHub IHub, auth transport.AuthMethod) error {
	var (
		form = &GitRepository{}
		to   = &GitRepository{}
	)
	if err := form.SetURL(formRepositoryURL); err != nil {
		return err
	}
	to.SetHub(toHub)
	to.SetAuth(auth)
	if err := to.SetURL(toRepositoryURL); err != nil {
		return err
	}
	return MirrorPushRepository(form, to)
}
func FileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
