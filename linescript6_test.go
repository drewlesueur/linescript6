package linescript6

import (
"fmt"
// "testing"
)

func ExampleE_hello() {
	E(`
        say1 .hello
    `)

	// Output:
	// hello
}

func ExampleParseString_yo() {
	tokens := ParseString(`
        say1 .hello
    `, "")
    for _, t := range tokens {
        fmt.Println("token: " + toJson(t.Name))
    }

	// Output:
    // token: "\n"
    // token: "say1"
    // token: ".hello"
    // token: "\n"
}