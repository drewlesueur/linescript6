package linescript6


import (
    "fmt"
    "strconv"
    "strings"
    "time"
    "log"
    "encoding/json"
)

type Token struct {
    Name string
    IsString bool
    Tokens []Token
    Action func(*State) *State
    SourceIndex int
    Source string
    Filename string
}

func ShowTokens(indent string, tokens []Token) string {
    str := ""
    for _, t := range tokens {
        if len(t.Tokens) > 0 {
            str += ShowTokens(indent + "    ", t.Tokens)
        } else {
            str += fmt.Sprintf("%s%q\n", indent, t.Name)
        }
    }
    return str
}

type State struct {
	I             int
	Code          []Token
	Vals          *List
	Vars          *Record
	LexicalParent *State
	CallingParent *State
	CurFuncInfo   *CurFuncInfo
	InCurrentCall bool
	NewlineSpot   int
	OnEndInfo     *OnEndInfo
	CallbacksCh chan Callback
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
type Callback struct {
	State        *State
	ReturnValues *List
	Vars *Record
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
	name := ""
	var funcToken func (s *State) *State = nil
	startToken := -1
	i := 0
	isString := false
	for i = i; i < len(src); i++ {
		chr := src[i]

		// time.Sleep(1 * time.Millisecond)
		// fmt.Println("    reading", i, string(chr), len(s.Code))
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
			case '(', '{', '[':
			    tokenStack = append(tokenStack, tokens)
			    tokens = []Token{}
			    continue
			case ')':
	 	    	localTokens := tokens
	 	    	localTokens = append(localTokens, Token{
			    	Action: immediates["\n"],
			    	Name: "\n",
			    	SourceIndex: i,
			    	Source: src,
			    	IsString: isString,
				})
	 	    	parentTokens := tokenStack[len(tokenStack)-1]
				parentTokens = append(parentTokens, Token{
	 	      		SourceIndex: i,
	 	      		Source: src,
	 	      		Tokens: localTokens,
	 	      		Name: "()",
	 	      		Action: func(s *State) *State {
  	     		        s.Push(localTokens)
  	     		        return s
	 	      		},
	 		    })
			    tokens = parentTokens
			    tokenStack = tokenStack[0 : len(tokenStack)-1]
			    continue
			case ']':
	 	    	localTokens := tokens
	 	    	localTokens = append(localTokens, Token{
			    	Action: immediates["\n"],
			    	Name: "\n",
			    	SourceIndex: i,
			    	Source: src,
			    	IsString: isString,
				})
	 	    	parentTokens := tokenStack[len(tokenStack)-1]
				parentTokens = append(parentTokens, Token{
	 	      		SourceIndex: i,
	 	      		Source: src,
	 	      		Tokens: localTokens,
	 	      		Name: "]",
	 	      		Action: func(s *State) *State {
	 	      		    vals := s.Vals
	 	      		    s.Vals = NewList()
	 	      		    s.Code = localTokens
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
	 	    	localTokens := tokens
	 	    	localTokens = append(localTokens, Token{
			    	Action: immediates["\n"],
			    	Name: "\n",
			    	SourceIndex: i,
			    	Source: src,
			    	IsString: isString,
				})
	 	    	parentTokens := tokenStack[len(tokenStack)-1]
				parentTokens = append(parentTokens, Token{
	 	      		SourceIndex: i,
	 	      		Source: src,
	 	      		Tokens: localTokens,
	 	      		Name: "}",
	 	      		Action: func(s *State) *State {
	 	      		    vals := s.Vals
	 	      		    s.Vals = NewList()
	 	      		    s.Code = localTokens
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
			default:
				parseState = "in"
				startToken = i
			}
		case "in":
			switch chr {
			case ' ', '\t', '\n', ';', ',', '(', ')', '{', '}', '[', ']':
                parseState = "out"
				name = src[startToken:i]
				i--
				if chr == ' ' || chr == '\t' {
				    i++
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
					name := name
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
		if name != "" {
			tokens = append(tokens, Token{
			    Action: funcToken,
			    Name: name,
			    SourceIndex: i,
			    Source: src,
			    IsString: isString,
			})
			name = ""
			isString = false
			funcToken = nil
		}
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
		s.InCurrentCall = true
		return s
		// if s.CurFuncInfo == nil {
		// 	return s
		// }
		// f := s.CurFuncInfo.Func
		// p := s.CurFuncInfo.Parent
		// s.CurFuncInfo = p
	 //    s.InCurrentCall = true
		// newS := f(s)
		// return newS
	},
	";": func(s *State) *State {
		s.InCurrentCall = true
		return s
		// if s.CurFuncInfo == nil {
		// 	return s
		// }
		// f := s.CurFuncInfo.Func
		// p := s.CurFuncInfo.Parent
		// s.CurFuncInfo = p
	 //    s.InCurrentCall = true
		// newS := f(s)
		// return newS
	},
	",": func(s *State) *State {
		if s.CurFuncInfo == nil {
			return s
		}
		f := s.CurFuncInfo.Func
		p := s.CurFuncInfo.Parent
		s.CurFuncInfo = p
		newS := f(s)
		return newS
	},
	// TODO: this is not hit.
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
		vI := s.Pop()
		v := vI.([]Token)
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
	GlobalState = NewGlobalState()
}

func E(sources ...any) {
	GlobalState.E(sources...)
}

func NewGlobalState() *State {
	s := &State{
		Vals:    NewList(),
		Vars:    NewRecord(), // since it's global, we reuse global vars
		CallbacksCh: make(chan Callback),
	}
	go s.Chug()
	return s
}

func (state *State) Chug() {
    var origState = state
	var newState *State
	for {
        newState = nil
        if state == nil {
            callback, ok := <-origState.CallbacksCh
            if !ok {
                break
            }
            state = callback.State
            
            if callback.ReturnValues != nil {
                for _, v := range callback.ReturnValues.TheSlice {
                    state.Push(v)
                }
            }
            if callback.Vars != nil {
                for _, k := range callback.Vars.Keys {
                    state.Vars.Set(k, callback.Vars.Get(k))
                }
            }
            state.NewlineSpot = state.Vals.Length()
            continue // you may be fine to not continue, cuz this is at the start
        }

        if state.InCurrentCall {
		    if state.CurFuncInfo == nil {
		    	state.InCurrentCall = false
			} else {
				f := state.CurFuncInfo.Func
				p := state.CurFuncInfo.Parent
				state.CurFuncInfo = p
				newState = f(state)
				state = newState
			}
			continue
        }

        if state.I >= len(state.Code) {
            o := state.OnEndInfo
            if o != nil {
                newState = o.OnEnd(state)
                state.OnEndInfo = o.Parent
            }
        } else {
            t := state.Code[state.I]
            log.Println("token:", t.Name)
            newState = t.Action(state)
        }

        state.I++
        state = newState
	}
}

func (s *State) E(code ...any) *State {
	filename := "__evaled_" + strconv.Itoa(int(time.Now().UnixNano()))
	tokens := Tokenize(code, filename)
	s.R(tokens)
	return s
}

func (s *State) R(tokens []Token) *State {
	doneCh := make(chan int)
	freshState := &State{
		I:             0,
		Code:          tokens,
		Vals:          s.Vals,
		Vars:          s.Vars,
		LexicalParent: s,
		CallingParent: nil,
		CallbacksCh: s.CallbacksCh,
		OnEndInfo: &OnEndInfo{
		    OnEnd: func (s *State) *State {
        		close(doneCh)
		        return nil
		    },
		},
	}
	s.AddCallback(Callback{
	    State: freshState,
	})
	<-doneCh
	return freshState
}

func (state *State) FindParentAndValue(varName string) (*State, any) {
	scopesUp := 0
	for state != nil {
		time.Sleep(500 * time.Millisecond)
		log.Println("going up scope after 500ms", varName)

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

func (state *State) AddCallback(callback Callback) {
	go func() {
		state.CallbacksCh <- callback
	}()
}

func toJson(v any) string {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(b)
}