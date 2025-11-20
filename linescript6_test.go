package linescript6

import (
    "fmt"
// "testing"
)

func ExampleParseString_block() {
	tokens := ParseString(`
        (say1 .hello)
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // {
    //     "say1"
    //     ".hello"
    //     "\n"
    // }
    // "\n"
}

func ExampleParseString_a() {
	tokens := ParseString(`
        say1 .hello
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // "say1"
    // ".hello"
    // "\n"
}
func ExampleParseString_def() {
	tokens := ParseString(`
        def foo a b
           a + b
        end
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // "def"
    // "foo"
    // "a"
    // "b"
    // {
    //     "a"
    //     "+"
    //     "b"
    //     "\n"
    // }
    // "\n"
}

func ExampleParseString_indent() {
	tokens := ParseString(`
        hello
           a b c
           d e f
        end ok
        yo
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // "hello"
    // {
    //     "a"
    //     "b"
    //     "c"
    //     "\n"
    //     "d"
    //     "e"
    //     "f"
    //     "\n"
    // }
    // "ok"
    // "\n"
    // "yo"
    // "\n"
}
func ExampleParseString_indent2() {
	tokens := ParseString(`
        hello
           a b
           d e
              more stuff
              here
           end
        end ok
        yo
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // "hello"
    // {
    //     "a"
    //     "b"
    //     "\n"
    //     "d"
    //     "e"
    //     {
    //         "more"
    //         "stuff"
    //         "\n"
    //         "here"
    //         "\n"
    //     }
    //     "\n"
    // }
    // "ok"
    // "\n"
    // "yo"
    // "\n"
}

func ExampleParseString_block3() {
	tokens := ParseString(`
        (say1 .hello) yo
    `, "")
    fmt.Println(ShowTokens(tokens))

    // Output:
    // {
    //     "say1"
    //     ".hello"
    //     "\n"
    // }
    // "yo"
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
    // token: "say1"
    // token: ".hello"
    // token: "\n"
}
func ExampleE_var() {
	E(`
        var .name .Drew
        say1 name
    `)

	// Output:
	// Drew
}