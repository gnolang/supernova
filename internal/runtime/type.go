package runtime

type Type string

const (
	RealmDeployment   Type = "REALM_DEPLOYMENT"
	PackageDeployment Type = "PACKAGE_DEPLOYMENT"
	RealmCall         Type = "REALM_CALL"
	unknown           Type = "UNKNOWN"
)

// IsRuntime checks if the passed in runtime
// is a supported runtime type
func IsRuntime(runtime Type) bool {
	return runtime == RealmCall ||
		runtime == RealmDeployment ||
		runtime == PackageDeployment
}

// String returns a string representation
// of the runtime type
func (r Type) String() string {
	switch r {
	case RealmDeployment:
		return string(RealmDeployment)
	case PackageDeployment:
		return string(PackageDeployment)
	case RealmCall:
		return string(RealmCall)
	default:
		return string(unknown)
	}
}
