module github.com/cyberark/conjur-k8s-csi-provider

go 1.22.7

toolchain go1.23.4

require (
	github.com/cyberark/conjur-api-go v0.11.1 // version will be ignored by auto release process
	github.com/cyberark/conjur-authn-k8s-client v0.26.1 // version will be ignored by auto release process
	github.com/hashicorp/go-version v1.6.0
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.63.2
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/apimachinery v0.30.8
	k8s.io/client-go v0.30.0
	sigs.k8s.io/secrets-store-csi-driver v1.4.7
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	google.golang.org/protobuf v1.36.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.30.8 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// Security fixes to ensure we don't have old vulnerable packages in our
// dependency tree. We're often not vulnerable, but removing them to ensure
// we never end up selecting them when other dependencies change.
// Only put specific versions on the left side of the =>
// so we don't downgrade future versions unintentionally.

replace golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20190620200207-3b0461eec859 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20200226121028-0de0cce0169b => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20201021035429-f5854403a974 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.0.0-20220722155237-a158d28d115b => golang.org/x/net v0.33.0

replace golang.org/x/net v0.6.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.10.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.15.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.21.0 => golang.org/x/net v0.33.0

replace golang.org/x/net v0.25.0 => golang.org/x/net v0.33.0

// Automated release process replaces
// DO NOT EDIT: CHANGES TO THESE 2 LINES WILL BREAK AUTOMATED RELEASES
replace github.com/cyberark/conjur-api-go => github.com/cyberark/conjur-api-go v0.12.9

replace github.com/cyberark/conjur-authn-k8s-client => github.com/cyberark/conjur-authn-k8s-client v0.26.4
