package linescript6

import (
    "fmt"
    "sync"
    "strconv"
    "time"
    "strings"
)

type CurFuncInfo struct {
	Func func(*State) *State
	Name string
	Spot int
	Parent *CurFuncInfo
}
type OnEndInfo struct {
	OnEnd func(*State) *State
	Parent *OnEndInfo
}

type State struct {
    Filename string
    I int
    Code string
	ICache []*ICache
    Vals *List
	Vars  *Record
	LexicalParent *State
	CallingParent *State
	CurFuncInfo *CurFuncInfo
	NewlineSpot int
	Mu *sync.Mutex
	Counter *int
	OnEndInfo *OnEndInfo
}

type TokenCacheValue struct {
    I int
    Name string
    TokenFunc func(*State) *State
}
type ICache struct {
    GoUp *int
    FindMatching *FindMatchingResult
    CachedToken *TokenCacheValue
}
type FindMatchingResult struct {
	Match  string
	I      int
	Indent string
}

func New() *State {
	counter := 0
	return &State{
		Vals:          NewList(),
		Vars:          NewRecord(), // since it's global, we reuse global vars
		Mu:            &sync.Mutex{},
		Counter: &counter,
	}
}
func (s *State) E(code string) *State {
    *s.Counter = *s.Counter + 1
	freshState := &State{
		Filename: "__evaled_" + strconv.Itoa(*s.Counter),
		I:             0,
		Code:          code,
		Vals:          s.Vals,
		Vars:          s.Vars,
		LexicalParent: s,
		CallingParent: nil,
		Mu:            s.Mu,
		Counter: s.Counter,
	}
	state := freshState
	for {
        // time.Sleep(500 * time.Millisecond)
        // fmt.Println("getting next token")

	    if state == nil {
	        break
	    }
	    // if state.I >= len(state.Code) {
	    //     break
	    // }
	    
	    name, tokenFunc := state.GetNextToken()
	    _ = name
	    // fmt.Println("got token", ToJsonF(name))
	    
	    if tokenFunc != nil {
	    	state = tokenFunc(state)
	    }
	}
	return freshState
}


var immediates = map[string]func(*State) *State {
    "\n": func(s *State) *State {
        for {
            // time.Sleep(500 * time.Millisecond)
            // fmt.Println("calling func", s.CurFuncName)

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

var builtins = map[string]func(*State) *State {
    "say1": func(s *State) *State {
        v := s.PopVal()
        fmt.Println(v)
        return s
    },
    "upper": func(s *State) *State {
        v := s.PopVal()
        s.Push(strings.ToUpper(toStringInternal(v)))
        return s
    },
    "lower": func(s *State) *State {
        v := s.PopVal()
        s.Push(strings.ToLower(toStringInternal(v)))
        return s
    },
}

func (s *State) GetNextToken() (string, func (*State) *State) {
    parseState := "out"
    name := "end"
    funcToken := immediates["end"]
    startToken := -1
    i := s.I
loop:
    for i = s.I; i < len(s.Code); i++ {
        // fmt.Println("    reading", i, len(s.Code))
        if i >= len(s.Code) {
            name = "end"
            funcToken = immediates[name]
            break loop
        }

        chr := s.Code[i]

        // time.Sleep(1 * time.Millisecond)
        // fmt.Println("    reading", i, string(chr), len(s.Code))

        switch parseState {
        case "out":
            switch chr {
            case ' ', '\t':
            case '\n', '(', ')', '{', '}', '[', ']', ';', ',':
                name = string(chr)
                if immediate, ok := immediates[name]; ok {
                    funcToken = immediate
                    i++
                    break loop
                }
            default:
                parseState = "in"
                startToken = i
            }
        case "in":
            switch chr {
            case ' ', '\t', '\n', '(', ')', '{', '}', '[', ']', ';', ',':
                name = s.Code[startToken:i]
                if chr == ' ' || chr == '\t' {
                    i++
                }
                if immediate, ok := immediates[name]; ok {
                    funcToken = immediate
                    i++
                    break loop
                }
                if builtin, ok := builtins[name]; ok {
                    funcToken = func(s *State) *State {
                        cfi := &CurFuncInfo{
                            Func:builtin,
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
                    break loop
                }
                if immediate, ok := immediates[name]; ok {
                    funcToken = immediate
                    break loop
                }

				if f, err := strconv.ParseFloat(name, 64); err == nil {
                    funcToken = func(s *State) *State {
                        s.Push(f)
                        return s
                    }
                    break loop
                }

                if len(name) >= 1 && name[0] == '.' {
                    funcToken = func(s *State) *State {
                        s.Push(name[1:])
                        return s
                    }
                    break loop
                }

                _, v := s.findParentAndValue(name)
                switch v.(type) {
                case *Func:
                    funcToken = func(s *State) *State {
                        v := s.Get(name).(*Func)
                        cfi := &CurFuncInfo{
                            Func: func (s *State) *State {
                                newState := &State{
                                    Filename: v.Filename,
                                    I: v.I,
                                    Code: v.Code,
                                    ICache: v.ICache,
                                    Vals: s.Vals,
	                                Vars: NewRecord(),
	                                LexicalParent: v.LexicalParent,
	                                CallingParent: s,
	                                CurFuncInfo: nil,
	                                NewlineSpot: s.NewlineSpot,
	                                Mu: s.Mu,
	                                Counter: s.Counter,
	                                OnEndInfo: nil,
                                }
  
		                        for i := len(v.Params) - 1; i >= 0; i-- {
		                        	param := v.Params[i]
		                        	newState.Vars.Set(param, s.PopVal())
		                        }
                                return newState
                            },
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
                default:
                    funcToken = func(s *State) *State {
                        s.Push(Token{Name: name})
                        return s
                    }
                }
                break loop
            }
        }
    }

    s.I = i
    return name, funcToken

}

type Token struct {
    Name string
}

var GlobalState *State

func init() {
	GlobalState = New()
}

func E(code string) {
    GlobalState.E(code)
}
func (s *State) GetVal(val any) any {
    switch val := val.(type) {
    case Token:
        return s.Get(val.Name)
    default:
        return val
    }
}

func (state *State) Get(varName string) any {
	parent, v := state.findParentAndValue(varName)
	if parent == nil {
		panic(fmt.Sprintf("var not found: %q", varName))
	}
	return v
}

func (state *State) findParentAndValue(varName string) (*State, any) {
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

func (s *State) Push(v any) {
	s.Vals.Push(v)
}
func (s *State) Pop() any {
	return s.Vals.Pop()
}
func (s *State) PopVal() any {
	 v := s.Vals.Pop()
	 return s.GetVal(v)
}

func (state *State) Let(varName string, v any) {
	parent, _ := state.findParentAndValue(varName)
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


type Func struct {
	Filename string
	I        int
	Code         string
	ICache []*ICache
	Params            []string
	LexicalParent     *State
	Name              string
}

