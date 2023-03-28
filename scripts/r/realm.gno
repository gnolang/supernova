package runtime

var greeting string

func init() {
	greeting = "Hello"
}

// SayHello says hello to the specified name, using
// the saved greeting
func SayHello(name string) string {
	return greeting + " " + name + "!"
}
