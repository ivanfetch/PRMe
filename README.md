# Pull Request Me

Pull Request Me (PR-me) creates a pull request for the entire content of a Github repository. This is useful to solicit review comments on an entire project.

## Sample Usage

* Set the `GH_TOKEN` environment variable to a [Github personal access token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) that has the `repo` scope; permission.

```
$ pr-me UserName/RepositoryName
Created pull request at https://github.com/UserName/RepositoryName/pulls/1
```

## How It Works

PR-me creates an orphaned branch with no commit history, and uses that branch as the base for a pull request. This allows the pull request to include all content present on the primary (typically the main or master) branch.
