package github

import (
	"context"
	"time"

	"github.com/dedecms/snake"
	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

var ctx = context.Background()
var client *github.Client

var owner, repo string

type Commit struct {
	Filename, Message string
}
type Commits struct {
	NewCommits []*Commit
}

func (c *Commits) Add(commit *Commit) *Commits {
	c.NewCommits = append(c.NewCommits, commit)
	return c
}

func init() {
	owner = "dedecms"
	repo = "5.7"
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "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"},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client = github.NewClient(httpClient)
}

func GetCommits(opts *github.CommitsListOptions) []*github.RepositoryCommit {
	if repos, _, err := client.Repositories.ListCommits(ctx, owner, repo, opts); err == nil {
		return repos
	}
	return nil
}

func GetCommit(sha string) *github.RepositoryCommit {
	if repos, _, err := client.Repositories.GetCommit(ctx, owner, repo, sha); err == nil {
		return repos
	}
	return nil
}

func GetTags() []*github.RepositoryTag {
	if repos, _, err := client.Repositories.ListTags(ctx, owner, repo, nil); err == nil {
		return repos
	}
	return nil
}

func GetNewTagSHA() string {
	if tags := GetTags(); tags != nil {
		for _, v := range GetTags() {
			return *v.Commit.SHA
		}
	}

	return ""
}

func GetNewCommit() *Commits {

	newCommit := new(Commits)
	tagCommit := GetCommit(GetNewTagSHA())
	commits := GetCommits(&github.CommitsListOptions{
		Until: time.Now(),
		Since: tagCommit.Commit.Committer.Date.Add(time.Second),
	})

	for _, v := range commits {
		commit := GetCommit(v.GetSHA())
		for _, file := range commit.Files {
			if file.GetStatus() == "modified" {
				newCommit.Add(&Commit{
					Filename: file.GetFilename(),
					Message:  snake.String(v.Commit.GetMessage()).GetOneLine(),
				})

			}
		}
	}

	return newCommit
}
