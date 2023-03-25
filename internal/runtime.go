package internal

type RuntimeType string

const (
	RealmDeployment   RuntimeType = "REALM_DEPLOYMENT"
	PackageDeployment RuntimeType = "PACKAGE_DEPLOYMENT"
	RealmCall         RuntimeType = "REALM_CALL"
)

func isRuntime(runtime RuntimeType) bool {
	return runtime == RealmCall ||
		runtime == RealmDeployment ||
		runtime == PackageDeployment
}

func (r RuntimeType) String() string {
	switch r {
	case RealmDeployment:
		return string(RealmDeployment)
	case PackageDeployment:
		return string(PackageDeployment)
	case RealmCall:
		return string(RealmCall)
	default:
		return "unknown runtime type"
	}
}
