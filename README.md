# meo-repo-manager

A CLI utility to automate GitHub repository creation within a specific organization, enforcing naming conventions and managing user access.

## Features

- **Automated Repository Creation**: Creates repositories in a configured GitHub organization.
- **Naming Convention Enforcement**: Enforces a configured prefix on repository names.
- **User Management**: Automatically adds a configured list of users as collaborators.
- **Interactive CLI**: Easy-to-use interactive prompts.

## Prerequisites

- A GitHub Personal Access Token (PAT) with `repo` scope.

## Configuration

The tool requires a configuration file (YAML format). By default, it looks for `config.yaml` in the same directory as the binary.

You can also specify a custom config file path using the `--config` flag.

### Example `config.yaml`

A template file `config.example.yaml` is provided in the repository.

```yaml
organization: "your-org"
prefix: "prefix-"
users:
  - "user1"
  - "user2"
```

- `organization`: The GitHub organization ID where repositories will be created.
- `prefix`: A string prefix that will be enforced on all repository names.
- `users`: A list of GitHub usernames to add as collaborators.

## Usage

1.  **Set your GitHub Token**:

    The tool reads the GitHub token from the `GITHUB_TOKEN` environment variable.

    ```bash
    export GITHUB_TOKEN=your_github_pat
    ```

    Alternatively, you can pass the token using the `--github-token` flag:

    ```bash
    ./meo-repo-manager create --github-token your_github_pat
    ```

2.  **Run the `create` command**:

    ```bash
    ./meo-repo-manager create
    ```

    Follow the interactive prompts to name your repository and select users.

3.  **Run the `manage` command**:

    To manage an existing repository (add/remove users):

    ```bash
    ./meo-repo-manager manage
    ```

    You can also specify the repository directly:

    ```bash
    ./meo-repo-manager manage --repo my-repo-name
    ```

4.  **Using a custom config file**:

    ```bash
    ./meo-repo-manager create --config /path/to/my-config.yaml
    ```

## Build

```bash
go build -o meo-repo-manager main.go
```