package gh

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient() *Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Warning: GITHUB_TOKEN environment variable is not set.")
		return &Client{client: github.NewClient(nil), ctx: context.Background()}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

func (c *Client) CreateRepository(org, name string) error {
	repo := &github.Repository{
		Name:    github.String(name),
		Private: github.Bool(true), // Default to private, maybe make configurable?
	}

	_, _, err := c.client.Repositories.Create(c.ctx, org, repo)
	return err
}

func (c *Client) AddCollaborator(owner, repo, user, permission string) error {
	opts := &github.RepositoryAddCollaboratorOptions{
		Permission: permission,
	}
	_, _, err := c.client.Repositories.AddCollaborator(c.ctx, owner, repo, user, opts)
	return err
}
