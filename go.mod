module github.com/cyberark/conjur-k8s-csi-provider

go 1.22.1

require (
	github.com/cyberark/conjur-api-go v0.11.1 // version will be ignored by auto release process
	github.com/cyberark/conjur-authn-k8s-client v0.26.1 // version will be ignored by auto release process
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.63.2
	gopkg.in/yaml.v3 v3.0.1
	sigs.k8s.io/secrets-store-csi-driver v1.4.2
)

require (
	github.com/alessio/shellescape v1.4.2 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/zalando/go-keyring v0.2.4 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// Automated release process replaces
// DO NOT EDIT: CHANGES TO THESE 2 LINES WILL BREAK AUTOMATED RELEASES
replace github.com/cyberark/conjur-api-go => github.com/cyberark/conjur-api-go latest

replace github.com/cyberark/conjur-authn-k8s-client => github.com/cyberark/conjur-authn-k8s-client latest
