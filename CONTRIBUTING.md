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
./bin/test_unit
```

This script generates a coverage profile in the `test/` directory. To generate a
more consumable HTML coverage profile:
```sh
go tool cover -html ./test/c.out -o ./test/c.html
open ./test/c.html
```

This project also includes end-to-end tests exercising core functionality. To run them:
```sh
./bin/test_e2e # KinD

./bin/test_e2e openshift {current-dev/oldest-dev/next-dev} # OpenShift
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

### Update the changelog and notices (if necessary)
1. Update the `CHANGELOG.md` file with the new version and the changes that are included in the release.
1. Update `NOTICES.txt`
    ```sh-session
    go install github.com/google/go-licenses@latest
    # Verify that dependencies fit into supported licenses types.
    # If there is new dependency having unsupported license, that license should be
    # included to notices.tpl file in order to get generated in NOTICES.txt.
    $(go env GOPATH)/bin/go-licenses check ./... \
      --allowed_licenses="MIT,ISC,Apache-2.0,BSD-3-Clause,BSD-2-Clause,MPL-2.0" \
      --ignore $(go list std | awk 'NR > 1 { printf(",") } { printf("%s",$0) } END { print "" }')
    # If no errors occur, proceed to generate updated NOTICES.txt
    $(go env GOPATH)/bin/go-licenses report ./... \
      --template notices.tpl \
      --ignore github.com/cyberark/conjur-k8s-csi-provider \
      --ignore $(go list std | awk 'NR > 1 { printf(",") } { printf("%s",$0) } END { print "" }') \
      > NOTICES.txt
    ```

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
