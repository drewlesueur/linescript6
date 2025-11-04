package main


import (
    "fmt"
)

type Token struct {
    Name string
    IsString bool
    TokensType string
    Tokens []Token
    Action func(*State) *State
    SourceIndex int
    Source string
    Filename string
}
type State struct {
	I             int
	Code          []Token
	Vals          *List
	Vars          *Record
	LexicalParent *State
	CallingParent *State
	CurFuncInfo   *CurFuncInfo
	NewlineSpot   int
	Mu            sync.Mutex
	OnEndInfo     *OnEndInfo
}
type CurFuncInfo struct {
	Func   func(*State) *State
	Name   string
	Spot   int
	Parent *CurFuncInfo
}
type OnEndInfo struct {
	OnEnd  func(*State) *State
	Parent *OnEndInfo
}

func Tokenize(sources []any, filename string) []Tokens {
    tokens := []Token{}
    for src := range sources {
        switch src := src.(type) {
        case string:
            tokens = append(tokens, ParseString(src, filename)...)
        case func(*State) *State:
			tokens = append(tokens, Token{
			    Action: src,
			    Filename: filename,
			})
		case func():
			tokens = append(tokens, Token{
			    Action: func(s *State) *State {
			        src()
			        return s
			    },
			    Filename: filename,
			})
		case func() any:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State {
			        v := src()
			        s.Push(v)
			        return s
			    },
			    Filename: filename,
			})
		case func(any) any:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State 			    Action: func(s *State) *State {
					s.Push(src(s.Pop()))
			    	return s
			    },
			    Filename: filename,
			})
		case func(any):
			tokens = append(tokens, Token{
			    Action: func(s *State) *State 			    Action: func(s *State) *State {
					src(s.Pop())
			    	return s
			    },
			    Filename: filename,
			})
		case func(string) string:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State
					s.Push(src(toStringInternal(s.Pop())))
			    	return s
			    },
			    Filename: filename,
			})
		case func(string, string) string:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State
					b := toStringInternal(s.Pop())
					a := toStringInternal(s.Pop())
					s.Push(src(a, b))
			    	return s
			    },
			    Filename: filename,
			})
		default:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State
					s.Push(src)
			    	return s
			    },
			    Filename: filename,
			})
        }
    }
    return tokens
}
// onEnd list (stack) of closures is the trick
// auto indent nested tokens
func ParseString(src, filename string) []Tokens {
	tokenStack := []Token{}
	tokens := []Token{}
	parseState := "out"
	name := "end"
	funcToken := immediates["end"]
	startToken := -1
	i := 0
	isString := false
	tokensType := "("
loop:
	for i = i; i < len(src); i++ {
		isString = false
		if i >= len(src) {
			name = "end"
			funcToken = immediates[name]
			break loop
		}

		chr := src[i]

		// time.Sleep(1 * time.Millisecond)
		// fmt.Println("    reading", i, string(chr), len(s.Code))
    theSwitch:
		switch chr {
		case '(', '{', '[':
		    tokenStack = append(tokenStack, Token{
		        Tokens: tokens,
		        TokensType: tokensType
		    })
		    tokensType = string(chr)
		    tokens = []Token{}
		    continue
		case ')':
	     	parentToken := tokenStack[len(tokenStack)-1]
			parentToken.Tokens = append(parentToken.Tokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: tokens,
	       		TokensType: tokensType,
	       		name: tokensType,
	       		Action: func(s *State) *State {
       		        s.Push(tokens)
       		        return s
	       		}
	 	    })
		    tokens = parentToken.Tokens
		    tokensType = parentToken.TokensType
		    tokenStack = tokenStack[0 : len(tokenStack)-1]
		    continue
		case ']':
	     	parentToken := tokenStack[len(tokenStack)-1]
			parentToken.Tokens = append(parentToken.Tokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: tokens,
	       		TokensType: tokensType,
	       		name: tokensType,
	       		Action: func(s *State) *State {
	       		    vals := s.Vals
	       		    s.Vals := NewList()
	       		    s.CodeStack = append(s.CodeStack, s.Code)
	       		    s.Code = tokens
	       		    s.OnEnd = func(s *State) *State {
	       		        
	       		    }
	       		}
	 	    })
		    tokens = parentToken.Tokens
		    tokensType = parentToken.TokensType
		    tokenStack = tokenStack[0 : len(tokenStack)-1]
		    continue
		case '}':
	     	parentToken := tokenStack[len(tokenStack)-1]
			parentToken.Tokens = append(parentToken.Tokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: tokens,
	       		TokensType: tokensType,
	       		name: tokensType,
	       		Action: func(s *State) *State {

       		        // TODO: could move this conditional out of closure
       		        if tokensType == "]" {

       		        } else if tokensType == "}" {

       		        }
	       		}
	 	    })
		    tokens = parentToken.Tokens
		    tokensType = parentToken.TokensType
		    tokenStack = tokenStack[0 : len(tokenStack)-1]
		    continue
		}
		switch parseState {
		case "out":
			switch chr {
			case ' ', '\t':
			case '\n', ';', ',':
				name = string(chr)
				if immediate, ok := immediates[name]; ok {
					funcToken = immediate
					// i++
					break theSwitch
				}
			default:
				parseState = "in"
				startToken = i
			}
		case "in":
			switch chr {
			case ' ', '\t', '\n', ';', ',':
                parseState = "out"
				name = src[startToken:i]
				if chr == ' ' || chr == '\t' {
					// i++
				}
				if immediate, ok := immediates[name]; ok {
					funcToken = immediate
					// i++
					break theSwitch
				}
				if builtin, ok := builtins[name]; ok {
					funcToken = func(s *State) *State {
						cfi := &CurFuncInfo{
							Func: builtin,
							Spot: s.Vals.Len(),
							Name: name,
						}
						if s.CurFuncInfo == nil {
							s.CurFuncInfo = cfi
						} else {
							cfi.Parent = s.CurFuncInfo
							s.CurFuncInfo = cfi
						}
						return s
					}
					break theSwitch
				}

				if f, err := strconv.ParseFloat(name, 64); err == nil {
					funcToken = func(s *State) *State {
						s.Push(f)
						return s
					}
					break theSwitch
				}

				if len(name) >= 1 && name[0] == '.' {
					isString = true
					funcToken = func(s *State) *State {
						s.Push(name[1:])
						return s
					}
					break theSwitch
				}

                // you could either do the closure
                // or just push 2 things to the tokens slice
				funcToken = func(s *State) *State {
					_, v := s.findParentAndValue(name)
					switch v.(type) {
					// case *Func:
					// 	v := s.Get(name).(*Func)
					// 	cfi := &CurFuncInfo{
					// 		Func: func(s *State) *State {
					// 			newState := &State{
					// 				Filename:      v.Filename,
					// 				I:             v.I,
					// 				Code:          v.Code,
					// 				Vals:          s.Vals,
					// 				Vars:          NewRecord(),
					// 				LexicalParent: v.LexicalParent,
					// 				CallingParent: s,
					// 				CurFuncInfo:   nil,
					// 				NewlineSpot:   s.NewlineSpot,
					// 				Mu:            s.Mu,
					// 				OnEndInfo:     nil,
					// 			}
                    //
					// 			for i := len(v.Params) - 1; i >= 0; i-- {
					// 				param := v.Params[i]
					// 				newState.Vars.Set(param, s.Pop())
					// 			}
					// 			return newState
					// 		},
					// 		Spot: s.Vals.Len(),
					// 		Name: name,
					// 	}
					// 	if s.CurFuncInfo == nil {
					// 		s.CurFuncInfo = cfi
					// 	} else {
					// 		cfi.Parent = s.CurFuncInfo
					// 		s.CurFuncInfo = cfi
					// 	}
					// 	return s
					default:
					    s.Push(v)
						return s
					}
				}
			}
		}
		tokens = append(tokens, Token{
		    Action: funcToken,
		    Name: name,
		    I: i,
		    IsString: isString,
		})
	}
    return tokens

}

// Slice is for 1 indexed slice
func Slice(state *State) *State {
	endInt := int(state.Pop().(float64))
	startInt := int(state.Pop().(float64))
	s := state.Pop()
	switch s := s.(type) {
	case *List:
		state.Push(s.Slice(startInt, endInt))
		return state
	case string:
		if len(s) == 0 {
			state.Push("")
			return state
		}
		if startInt < 0 {
			startInt = len(s) + startInt + 1
		}
		if startInt <= 0 {
			startInt = 1
		}
		if startInt > len(s) {
			state.Push("")
			return state
		}
		if endInt < 0 {
			endInt = len(s) + endInt + 1
		}
		if endInt <= 0 {
			state.Push("")
			return state
		}
		if endInt > len(s) {
			endInt = len(s)
		}
		if startInt > endInt {
			state.Push("")
			return state
		}
		state.Push(s[startInt-1 : endInt])
		return state
	}
	state.Push(nil)
	return state
}

var immediates = map[string]func(*State) *State{
	"\n": func(s *State) *State {
		for {
			if s.CurFuncInfo == nil {
				break
			}
			newS := s.CurFuncInfo.Func(s)
			if newS == s {
				newS.CurFuncInfo = newS.CurFuncInfo.Parent
			}
			s = newS
		}
		return s
	},
	";": func(s *State) *State {
		for {
			if s.CurFuncInfo == nil {
				break
			}
			newS := s.CurFuncInfo.Func(s)
			// like a bee that only stings once
			if newS == s {
				newS.CurFuncInfo = newS.CurFuncInfo.Parent
			}
			s = newS
		}
		return s
	},
	",": func(s *State) *State {
		if s.CurFuncInfo == nil {
			return s
		}
		newS := s.CurFuncInfo.Func(s)
		// like a bee that only stings once
		if newS == s {
			newS.CurFuncInfo = newS.CurFuncInfo.Parent
		}
		s = newS
		return s
	},
	"end": func(s *State) *State {
		if s.OnEndInfo != nil {
			newS := s.OnEndInfo.OnEnd(s)
			if newS == s {
				newS.OnEndInfo = newS.OnEndInfo.Parent
			}
			return newS
		}
		s = s.CallingParent
		return s
	},
}

var builtins = map[string]func(*State) *State{
	"say1": func(s *State) *State {
		v := s.Pop()
		fmt.Println(v)
		return s
	},
	"upper": func(s *State) *State {
		v := s.Pop()
		s.Push(strings.ToUpper(toStringInternal(v)))
		return s
	},
	"lower": func(s *State) *State {
		v := s.Pop()
		s.Push(strings.ToLower(toStringInternal(v)))
		return s
	},
}

func (s *State) Push(v any) {
	s.Vals.Push(v)
}
func (s *State) Pop() any {
	return s.Vals.Pop()
}