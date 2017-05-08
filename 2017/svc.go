package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"

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
	cl := githubClient(nethttp.OperationName("Repositories.List"))
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
			cl := githubClient(nethttp.OperationName("Repositories.ListBranches"))
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
	start := time.Now()
	var (
		owner string
		err   error
	)
	defer func() {
		errS := ""
		if err != nil {
			errS = err.Error()
		}
		log.Printf("request owner=%v duration=%v error=%q", owner, time.Since(start), errS)
	}()

	owner = strings.TrimPrefix(r.URL.Path, "/")
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
	closer, err := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
		},
	}.InitGlobalTracer("mostBranchesSvc")
	if err != nil {
		log.Println("Could not initialize jaeger tracer:", err)
	} else {
		defer closer.Close()
	}

	handler := nethttp.Middleware(opentracing.GlobalTracer(), http.HandlerFunc(handler))
	server := &http.Server{Addr: ":8080", Handler: handler}
	log.Println("Listening on", server.Addr)
	err = server.ListenAndServe()
	log.Fatal(err)
}

func githubClient(options ...nethttp.ClientOption) *github.Client {
	// We use http.RoundTrip to trace calls done by go-github, rather than
	// instrumentating each call site.
	t := &tracingTransport{
		Options: options,
		RoundTripper: &nethttp.Transport{
			RoundTripper: githubTransport,
		},
	}
	return github.NewClient(&http.Client{Transport: t})
}

type tracingTransport struct {
	Options []nethttp.ClientOption
	http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req, t.Options...)
	defer ht.Finish()
	return t.RoundTripper.RoundTrip(req)
}
