package runtime

type Language string

const (
	French    Language = "french"
	Italian   Language = "italian"
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
