package linescript6


import (
    "fmt"
    "strconv"
    "strings"
    "time"
)

type Token struct {
    Name string
    IsString bool
    Tokens *[]Token
    Action func(*State) *State
    SourceIndex int
    Source string
    Filename string
}
type State struct {
	I             int
	Code          *[]Token
	Vals          *List
	Vars          *Record
	LexicalParent *State
	CallingParent *State
	CurFuncInfo   *CurFuncInfo
	InCurrentCall bool
	NewlineSpot   int
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

func Tokenize(sources []any, filename string) []Token {
    tokens := []Token{}
    for _, src := range sources {
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
			    Action: func(s *State) *State {
					s.Push(src(s.Pop()))
			    	return s
			    },
			    Filename: filename,
			})
		case func(any):
			tokens = append(tokens, Token{
			    Action: func(s *State) *State {
					src(s.Pop())
			    	return s
			    },
			    Filename: filename,
			})
		case func(string) string:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State {
					s.Push(src(toStringInternal(s.Pop())))
			    	return s
			    },
			    Filename: filename,
			})
		case func(string, string) string:
			tokens = append(tokens, Token{
			    Action: func(s *State) *State {
					b := toStringInternal(s.Pop())
					a := toStringInternal(s.Pop())
					s.Push(src(a, b))
			    	return s
			    },
			    Filename: filename,
			})
		default:
			tokens = append(tokens, Token{
			    Name: "customType",
			    Action: func(s *State) *State {
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
func ParseString(src, filename string) []Token {
	tokenStack := [][]Token{}
	tokens := []Token{}
	parseState := "out"
	name := "end"
	funcToken := immediates["end"]
	startToken := -1
	i := 0
	isString := false
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
		switch chr {
		case '(', '{', '[':
		    tokenStack = append(tokenStack, tokens)
		    tokens = []Token{}
		    continue
		case ')':
	     	parentTokens := tokenStack[len(tokenStack)-1]
			parentTokens = append(parentTokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: &tokens,
	       		Name: ")",
	       		Action: func(s *State) *State {
       		        s.Push(&tokens)
       		        return s
	       		},
	 	    })
		    tokens = parentTokens
		    tokenStack = tokenStack[0 : len(tokenStack)-1]
		    continue
		case ']':
	     	parentTokens := tokenStack[len(tokenStack)-1]
			parentTokens = append(parentTokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: &tokens,
	       		Name: "]",
	       		Action: func(s *State) *State {
	       		    vals := s.Vals
	       		    s.Vals = NewList()
	       		    s.Code = &tokens
	       		    s.OnEndInfo = &OnEndInfo{
	       		        OnEnd: func(s *State) *State {
	       		            myList := s.Vals
	       		            s.Vals = vals
	       		            s.Vals.Push(myList)
	       		            return s
	       		    	},
	       		    	Parent: s.OnEndInfo,
	       		    }
	       		    return s
	       		},
	 	    })
		    tokens = parentTokens
		    tokenStack = tokenStack[0 : len(tokenStack)-1]
		    continue
		case '}':
	     	parentTokens := tokenStack[len(tokenStack)-1]
			parentTokens = append(parentTokens, Token{
	       		SourceIndex: i,
	       		Source: src,
	       		Tokens: &tokens,
	       		Name: "}",
	       		Action: func(s *State) *State {
	       		    vals := s.Vals
	       		    s.Vals = NewList()
	       		    s.Code = &tokens
	       		    s.OnEndInfo = &OnEndInfo{
	       		        OnEnd: func(s *State) *State {
	       		            myList := s.Vals
	       		            myRecord := NewRecord()
	       		            for i := 0; i < myList.Length() - 1; i += 2 {
	       		                myRecord.Set(myList.Get(i+1).(string), myList.Get(i+2))
	       		            }
	       		            s.Vals = vals
	       		            s.Vals.Push(myRecord)
	       		            return s
	       		    	},
	       		    	Parent: s.OnEndInfo,
  	       		    }
       	            return s
	       		},
	 	    })
		    tokens = parentTokens
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
					break
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
					break
				}
				if builtin, ok := builtins[name]; ok {
					funcToken = func(s *State) *State {
						s.CurFuncInfo = &CurFuncInfo{
							Func: builtin,
							Spot: s.Vals.Len(),
							Name: name,
							Parent: s.CurFuncInfo,
						}
						return s
					}
					break
				}

				if f, err := strconv.ParseFloat(name, 64); err == nil {
					funcToken = func(s *State) *State {
						s.Push(f)
						return s
					}
					break
				}

				if len(name) >= 1 && name[0] == '.' {
					isString = true
					funcToken = func(s *State) *State {
						s.Push(name[1:])
						return s
					}
					break
				}

                // you could either do the closure
                // or just push 2 things to the tokens slice
				funcToken = func(s *State) *State {
					_, v := s.FindParentAndValue(name)
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
		    SourceIndex: i,
		    Source: src,
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
			} else {
			    s.InCurrentCall = true
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
			} else {
			    s.InCurrentCall = true
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
	"do": func(s *State) *State {
		v := s.Pop().(*[]Token)
		oldCode := s.Code
		oldI := s.I
		s.Code = v
		s.I = 0
        s.OnEndInfo = &OnEndInfo{
           OnEnd: func(s *State) *State {
               s.Code = oldCode
               s.I = oldI
               return s
           },
           Parent: s.OnEndInfo,
        }
		
		return s
	},
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
var GlobalState *State

func init() {
	GlobalState = New()
}

func E(sources ...any) {
	GlobalState.E(sources...)
}
func New() *State {
	return &State{
		Vals:    NewList(),
		Vars:    NewRecord(), // since it's global, we reuse global vars
	}
}

func (s *State) E(code ...any) *State {
	filename := "__evaled_" + strconv.Itoa(int(time.Now().UnixNano()))
	tokens := Tokenize(code, filename)
	s.R(tokens)
	return s
}

func (s *State) R(tokens []Token) *State {
	freshState := &State{
		I:             0,
		Code:          &tokens,
		Vals:          s.Vals,
		Vars:          s.Vars,
		LexicalParent: s,
		CallingParent: nil,
	}
	state := freshState
	origState := state
	origTokens := &tokens

	for {
		// time.Sleep(500 * time.Millisecond)
		// fmt.Println("getting next token")

		if state == nil {
			break
		}
		// if state.I >= len(state.Code) {
		//     break
		// }
        t := (*state.Code)[state.I]
        newState := t.Action(state)

        if newState == nil {
            // <- state.Ch
        } else if newState == origState && newState.Code == origTokens && newState.I >= len(*newState.Code)-1  {
            break
        }
        state.I++
        state = newState
	}
	return freshState
}

func (state *State) FindParentAndValue(varName string) (*State, any) {
	scopesUp := 0
	for state != nil {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("going up scope", varName)

		v, ok := state.Vars.GetHas(varName)
		if ok {
			return state, v
		}
		state = state.LexicalParent
		scopesUp++
	}
	return nil, nil
}
func (state *State) Let(varName string, v any) {
	parent, _ := state.FindParentAndValue(varName)
	if parent == nil {
		panic("var not found " + varName)
	}
	parent.Vars.Set(varName, v)
}
func (state *State) Var(varNameAny any, v any) {
	varName := varNameAny.(string)
	if state.Vars == nil {
		state.Vars = NewRecord()
	}
	state.Vars.Set(varName, v)
}
