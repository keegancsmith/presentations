package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/google/go-github/github"
)

var githubTransport = &github.UnauthenticatedRateLimitedTransport{
	ClientID:     os.Getenv("GITHUB_CLIENT"),
	ClientSecret: os.Getenv("GITHUB_SECRET"),
}

type repoBranchCount struct {
	Repo     string
	Branches int
}

func mostBranches(ctx context.Context, owner string) (*repoBranchCount, error) {
	cl := github.NewClient(githubTransport.Client())
	repos, _, err := cl.Repositories.List(ctx, owner, nil)
	if err != nil {
		return nil, err
	}
	var (
		g   errgroup.Group
		mu  sync.Mutex
		max repoBranchCount
	)
	for _, repo := range repos {
		repo := repo
		g.Go(func() error {
			branches, _, err := cl.Repositories.ListBranches(ctx, *repo.Owner.Login, *repo.Name, nil)
			if err != nil {
				return err
			}
			mu.Lock()
			if len(branches) > max.Branches {
				max.Branches = len(branches)
				max.Repo = *repo.FullName
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &max, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	owner := strings.TrimPrefix(r.URL.Path, "/")
	max, err := mostBranches(r.Context(), owner)
	if err != nil {
		code := http.StatusInternalServerError
		if gerr, ok := err.(*github.ErrorResponse); ok {
			code = gerr.Response.StatusCode
		}
		http.Error(w, err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, err = fmt.Fprintf(w, "github.com/%s has the most branches (%d) for %s\n", max.Repo, max.Branches, owner)
	if err != nil {
		log.Println("failed to respond to client", err)
	}
}

func main() {
	handler := http.HandlerFunc(handler)
	server := &http.Server{Addr: ":8080", Handler: handler}
	log.Println("Listening on", server.Addr)
	err := server.ListenAndServe()
	log.Fatal(err)
}
