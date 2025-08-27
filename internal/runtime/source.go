package runtime

const (
	realmBody = `package runtime

var greeting string

func init() {
	greeting = "Hello"
}

// SayHello says hello to the specified name, using
// the saved greeting
func SayHello(name string) string {
	return greeting + " " + name + "!"
}
`
	packageBody = `package runtime

type Language string

const (
	French    Language = "french"
	Italian   Language = "italian"
	Spanish   Language = "spanish"
	Hindi     Language = "hindi"
	Bulgarian Language = "bulgarian"
	Serbian   Language = "serbian"
)

// GetGreeting generates a greeting in
// the specified language
func GetGreeting(language Language) string {
	switch language {
	case French:
		return "Bonjour"
	case Italian:
		return "Ciao"
	case Spanish:
		return "Hola"
	case Hindi:
		return "नमस्ते"
	case Bulgarian:
		return "Здравейте"
	case Serbian:
		return "Здраво"
	default:
		return "Hello"
	}
}
`
	gnomodBody = `
module = "gno.land/r/demo/runtime"
gno = "0.9"
`
)

const (
	packageName     = "runtime"
	realmFileName   = "realm.gno"
	packageFileName = "package.gno"
	gnomodFileName  = "gnomod.toml"
)
