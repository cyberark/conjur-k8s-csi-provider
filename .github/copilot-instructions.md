# Copilot Instructions

## go.mod: Automated Release Replace Directives

The following lines in `go.mod` are managed exclusively by the automated release process. **Never modify, remove, or suggest changes to these lines**, even when updating dependencies, bumping the Go directive, or resolving conflicts:

```
// Automated release process replaces
// DO NOT EDIT: CHANGES TO THESE 2 LINES WILL BREAK AUTOMATED RELEASES
replace github.com/cyberark/conjur-api-go => github.com/cyberark/conjur-api-go <version>

replace github.com/cyberark/conjur-authn-k8s-client => github.com/cyberark/conjur-authn-k8s-client <version>
```

The version pinned on the right-hand side of each `replace` directive is injected by CI. Any manual edit (including version bumps or reformatting) will break automated releases.
