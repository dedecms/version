package github

import (
	"context"
	"fmt"
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
func GetNewTarPackage() string {
	url, _, _ := client.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, nil, false)
	return url.String()
}

func GetUpdateList() *Commits {

	newCommit := new(Commits)

	tagCommit := GetCommit(GetNewTagSHA())
	commits := GetCommits(&github.CommitsListOptions{
		Until: time.Now(),
		Since: tagCommit.Commit.Committer.Date.Add(time.Second),
	})
	count := make(map[string]int)

	for _, v := range commits {
		commit := GetCommit(v.GetSHA())
		msg := v.Commit.GetMessage()
		msg = snake.String(msg).GetOneLine()

		if m := snake.String(msg).Extract(`\[(.*)\]([ |	]{0,})(.*)`, "$1"); len(m) == 1 {
			tag := snake.String(m[0]).Trim(" ").Trim("	").ToUpper().Get()
			count[tag]++
			if tag == "SU" {
				msg = fmt.Sprintf("[%s-%s-%d] 提高了DedeCMS的安全性，建议所有官方原版程序搭建的站点都进行安装", tag, time.Now().Format("20060102"), count[tag])
			} else {
				m = snake.String(msg).Extract(`\[(.*)\]([ |	]{0,})(.*)`, fmt.Sprintf("[%s-%s-%d] $3", tag, time.Now().Format("20060102"), count[tag]))
				msg = m[0]
			}
			for _, file := range commit.Files {
				if file.GetStatus() != "deleted" && snake.FS("./source").Add(file.GetFilename()).Exist() {
					newCommit.Add(&Commit{
						Filename: file.GetFilename(),
						Message:  msg,
					})

				}
			}
		}
	}

	return newCommit
}
