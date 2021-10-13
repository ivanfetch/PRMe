# Pull Request Me

Pull Request Me (PR-me) creates a pull request for the entire content of a Github repository. This is useful to solicit review comments on an entire project.

## Usage

* Set the `GH_TOKEN` environment variable to a [Github personal access token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) that has the `repo` scope; permission.
	* Note that the `repo` scope allows access to any repository that is available to your Github account - there currently is no more granular repository permission available.
* Have [Git](https://git-scm.com/downloads) installed.
	* Be sure Github SSH access to clone and push repositories work correctly, using URLs of the form `ssh://git@github.com/...`.
* Build this PR-me tool - after downloading or cloning this repository: `make build`

```
$ ./prme UserName/RepositoryName
A full pull request has been created at https://github.com/UserName/RepositoryName/pulls/1
```

* After the pull request has been merged, be sure to *manually merge* the base branch of the pull request into the default branch of your repository (typically `main` or `master).

Run `./prme -h` for additional options, including the default repository branch, pull request title and body (first comment), and names to be used for the pull request branches.

## How It Works

PR-me creates an orphaned branch with no commit history, as the base for a pull request. This allows the pull request to include all content present on the default branch of the repository (typically `main` or `master`).

This utility performs these steps to accomplish the above:

* Use the `git` command to clone the repository via SSH, and create two orphan branches as the base and head branches for the pull request. Remaining steps will use the Github API instead of the local `git` command.
* Merge the default branch (typically `main` or `master`) into the head pull request branch.
* Create a pull request using the empty orphan base branch, and the head branch which contains the same history as the default branch.


## Design Considerations

### Using Git

The git command is used in one area where the Github API cannot be used - creating an orphan branch with no files in the repository.

Unfortunately the [Github API to create a commit](https://docs.github.com/en/rest/reference/git#create-a-commit) does not support reliably creating a commit pointing at the [git empty-tree](https://stackoverflow.com/questions/9765453/is-gits-semi-secret-empty-tree-object-reliable-and-why-is-there-not-a-symbolic). The Github API call often returns an HTTP 500, with no HTTP body.

Github technical support responded that this specific (empty-tree) operation should never succeed, and that the 500 error is expected.

Here is an example command to create a commit for the empty-tree, using the Github API:

```
curl -v \
-X POST \
-H "Accept: application/vnd.github.v3+json" \
-H "Authorization: token ${GH_TOKEN}" \
https://api.github.com/repos/UserName/TestRepo/git/commits \
-d '{"message":"empty tree commit","tree":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}'
```

