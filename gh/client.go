package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
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

func (c *Client) Validate() error {
	_, _, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	return nil
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
