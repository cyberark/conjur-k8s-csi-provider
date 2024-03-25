# Contributing

For general contribution and community guidelines, please see the
[community repo](https://github.com/cyberark/community).

## Prerequisites

1. [git](https://git-scm.com/downloads) to manage source code
2. [Docker](https://docs.docker.com/engine/installation) to manage dependencies
   and runtime environments
3. [Go 1.22.1+](https://go.dev/doc/install) installed

## Testing

This project includes table-driven unit tests for each component. To run them:
```sh
./bin/test
```

This script generates a coverage profile in the `test/` directory. To generate a
more consumable HTML coverage profile:
```sh
go tool cover -html ./test/c.out -o ./test/c.html
open ./test/c.html
```

## Pull Request Workflow

1. [Fork the project](https://help.github.com/en/github/getting-started-with-github/fork-a-repo)
2. [Clone your fork](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/cloning-a-repository)
3. Make local changes to your fork by editing files
4. [Commit your changes](https://help.github.com/en/github/managing-files-in-a-repository/adding-a-file-to-a-repository-using-the-command-line)
5. [Push your local changes to the remote server](https://help.github.com/en/github/using-git/pushing-commits-to-a-remote-repository)
6. [Create new Pull Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)

From here your pull request will be reviewed and once you've responded to all
feedback it will be merged into the project. Congratulations, you're a
contributor!

## Releases

Releases should be created by maintainers only. To create a tag and release,
follow the instructions in this section.

### Pre-requisites

1. Review the git log and ensure the [changelog](CHANGELOG.md) contains all
   relevant recent changes with references to GitHub issues or PRs, if possible.
   Also ensure the latest unreleased version is accurate - our pipeline generates 
   a VERSION file based on the changelog, which is then used to assign the version
   of the release and any release artifacts.
1. Review the changes since the last tag, and if the dependencies have changed
   revise the [NOTICES](NOTICES.txt) to correctly capture the included
   dependencies and their licenses / copyrights.
1. Ensure that all documentation that needs to be written has been 
   written by TW, approved by PO/Engineer, and pushed to the forward-facing documentation.
1. Scan the project for vulnerabilities

### Release and Promote

1. Merging into main/master branches will automatically trigger a release. If successful, this release can be promoted at a later time.
1. Jenkins build parameters can be utilized to promote a successful release or manually trigger aditional releases as needed.
1. Reference the [internal automated release doc](https://github.com/conjurinc/docs/blob/master/reference/infrastructure/automated_releases.md#release-and-promotion-process) for releasing and promoting.

### Push Helm package

1. Every release build packages the CSI Provider Helm chart for us. The package can be found on the draft (or published) release for the relevant version.
1. Clone the repo [helm-charts](https://github.com/cyberark/helm-charts) and do the following:
    1. Move the Helm package file created in the previous step to the *docs* folder in the `helm-charts` repo.
    1. Go to the `helm-charts` repo root folder and execute the `reindex.sh` script file located there.
    1. Create a PR with those changes.
