package linescript6

import (
"fmt"
// "testing"
)

func ExampleParseString_a() {
	tokens := ParseString(`
        say1 .hello
    `, "")
    fmt.Println(ShowTokens("", tokens))

    // Output:
    // "\n"
    // "say1"
    // ".hello"
    // "\n"
}

func ExampleParseString_block() {
	tokens := ParseString(`
        (say1 .hello)
    `, "")
    fmt.Println(ShowTokens("", tokens))

    // Output:
    // "\n"
    //     "say1"
    //     ".hello"
    //     "\n"
    // "\n"
}
func ExampleParseString_block3() {
	tokens := ParseString(`
        (say1 .hello) do
    `, "")
    fmt.Println(ShowTokens("", tokens))

    // Output:
    // "\n"
    //     "say1"
    //     ".hello"
    //     "\n"
    // "do"
    // "\n"
}

// func ExampleE_fail() {
//     panic("fake fail")
// 	// Output:
// }


func ExampleE_block() {
	E(`
        (say1 .hello) do
    `)

	// Output:
	// hello
}

func ExampleE_hello() {
	E(`
        say1 .hello
    `)

	// Output:
	// hello
}


func ExampleE_helloUpper() {
	E(`
        say1 upper .hello
        .Cloud upper
        say1

        say1 .wallpaper upper

        .grape upper, say1

        .square upper; say1

    `)

	// Output:
	// HELLO
	// CLOUD
	// WALLPAPER
	// GRAPE
	// SQUARE
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