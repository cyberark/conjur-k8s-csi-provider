package logmessages

/*
	This go file centralizes log messages (in different levels) so we have them all in one place.

	Although having the names of the consts as the error code (i.e CAKC001) and not as a descriptive name (i.e WriteAccessTokenError)
	can reduce readability of the code that raises the error, we decided to do so for the following reasons:
		1.  Improves supportability – when we get this code in the log we can find it directly in the code without going
			through the “log_messages.go” file first
		2. Validates we don’t have error code duplications – If the code is only in the string then 2 errors can have the
			same code (which is something that a developer can easily miss). However, If they are in the object name
			then the compiler will not allow it.
*/

const CKCP001 string = "CKCP001 Kubernetes CSI Provider %s starting up..."
const CKCP002 string = "CKCP002 Invalid log level: %s. Defaulting to info"
const CKCP003 string = "CKCP003 CSI provider server failed: %v"
const CKCP004 string = "CKCP004 CSI provider health server failed: %v"
const CKCP005 string = "CKCP005 Failed to stop the CSI provider health server: %v"
const CKCP006 string = "CKCP006 Unsupported configuration version: %q"
const CKCP007 string = "CKCP007 Failed to unmarshal attribute %q: %w"
const CKCP008 string = "CKCP008 Missing serviceaccount token for audience %q"
const CKCP009 string = "CKCP009 Missing required Conjur config attributes: %q"
const CKCP010 string = "CKCP010 Attribute \"%s\" missing or empty"
const CKCP011 string = "CKCP011 Failed to unmarshal secrets spec: %w"
const CKCP012 string = "CKCP012 Failed to unmarshal file permissions: %w"
const CKCP013 string = "CKCP013 Failed to create configuration from mount request parameters: %w"
const CKCP014 string = "CKCP014 Can't append Conjur SSL cert"
const CKCP015 string = "CKCP015 Request failed with status code %d"
const CKCP016 string = "CKCP016 Failed to get Conjur secrets: %w"
const CKCP017 string = "CKCP017 Failed to unmarshal attributes: %w"
const CKCP018 string = "CKCP018 Creating and registering gRPC server..."
const CKCP019 string = "CKCP019 Using non-standard providers directory %s: Ensure this directory has been configured on your CSI Driver before proceeding"
const CKCP020 string = "CKCP020 Failed to start socket listener: %w"
const CKCP021 string = "CKCP021 Serving gRPC server on socket %s..."
const CKCP022 string = "CKCP022 Stopping gRPC server..."
const CKCP023 string = "CKCP023 gRPC server stopped."
const CKCP024 string = "CKCP024 Serving health server on port %d..."
const CKCP025 string = "CKCP025 Stopping health server..."
const CKCP026 string = "CKCP026 Health server stopped."
const CKCP030 string = "CKCP030 Failed to create Conjur client: %v"
const CKCP031 string = "CKCP031 Failed to retrieve batch secrets: %v"
const CKCP032 string = "CKCP032 Failed to parse request attributes: %v"
const CKCP033 string = "CKCP033 Failed to unmarshal YAML: %v"
const CKCP034 string = "CKCP034 Annotation \"%s\" missing or empty"
const CKCP035 string = "CKCP035 Failed to retrieve pod annotations: %v"
const CKCP036 string = "CKCP036 Creating Kubernetes client..."
const CKCP037 string = "CKCP037 Failed to load kubeconfig."
const CKCP038 string = "CKCP038 Failed to configure k8s client."
const CKCP039 string = "CKCP039 Failed to get pod \"%s\" in namespace \"%s\": %v"
const CKCP040 string = "CKCP040 Provided configuration version: %v"
const CKCP041 string = "CKCP041 Configuration version not provided. Defaulting to: %v"
const CKCP042 string = "CKCP042 Defining secrets in the SecretProviderClass is deprecated in v0.2.0 and greater. Please use the 'conjur.org/secrets' annotation in the pod spec."
