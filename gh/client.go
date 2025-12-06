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
		Name:    github.Ptr(name),
		Private: github.Ptr(true), // Default to private, maybe make configurable?
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

func (c *Client) GetRepository(org, name string) (*github.Repository, error) {
	repo, _, err := c.client.Repositories.Get(c.ctx, org, name)
	return repo, err
}

func (c *Client) ListRepositories(org, prefix string) ([]*github.Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Direction:   "desc",
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(c.ctx, org, opts)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if repo.Name != nil && (prefix == "" || (len(*repo.Name) >= len(prefix) && (*repo.Name)[0:len(prefix)] == prefix)) {
				allRepos = append(allRepos, repo)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepos, nil
}

func (c *Client) ListCollaborators(owner, repo string) ([]*github.User, error) {
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Affiliation: "all",
	}
	var allUsers []*github.User
	for {
		users, resp, err := c.client.Repositories.ListCollaborators(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allUsers, nil
}

func (c *Client) RemoveCollaborator(owner, repo, user string) error {
	_, err := c.client.Repositories.RemoveCollaborator(c.ctx, owner, repo, user)
	return err
}

func (c *Client) GetPermissionLevel(owner, repo, user string) (*github.RepositoryPermissionLevel, error) {
	perm, _, err := c.client.Repositories.GetPermissionLevel(c.ctx, owner, repo, user)
	return perm, err
}
