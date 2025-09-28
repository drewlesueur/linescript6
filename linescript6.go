package linescript6

import (
    "fmt"
    "sync"
)

type State struct {
    FileName
    I int
    Code string
    Vals *List
	Vars  *Record
	LexicalParent *State
	CallingParent *State
	CurFunc func(*State) *State
	CurFuncSpot int
	NewlineSpot int
	Mu sync.Mutex
	Counter int
	ICache []*ICache
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
	&State{
		Vals:          NewList(),
		Vars:          NewRecord(), // since it's global, we reuse global vars
		LexicalParent: s,
		Mu:            sync.Mutex,
	}
}
func (s *State) E(code state) *State {
    s.Counter++
	freshState := &State{
		FileName: "__evaled_" + strconv.Itoa(s.Counter),
		I:             0,
		Code:          code,
		Vals:          s.Vals,
		Vars:          s.Vars,
		LexicalParent: s,
		CallingParent: nil,
		Mu:            s.Mu,
	}
	state := freshState
	for {
	    if state == nil {
	        break
	    }
	    name, tokenFunc := state.GetNextToken()
	    _ = name
	    state = tokenFunc(state)
	}
	return freshState
}


var immediates = map[string]func(*State) *State {
    "\n": func(s *State) *State {
        for {
            if s.CurFunc == nil {
                break
            }
            s = s.CurFunc(s)
        }
        return s
    },
    
}
var builtins = map[string]func(*State) *State {
    "say1": func(s *State) *State {
        v := s.Pop()
        fmt.Println()
    },
}


func (s *State) GetNextToken() (string, func (*State) *State) {
    parseState := "out"
    
    name := ""
    var funcToken func(*State) *State
    startToken := -1
    i := s.I
loop:
    for i = s.I; i < len(s.Code); i++ {
        chr := s.Code[i]
        switch chr {
        case "\n", "(", ")", "{", "}", "[", "]", ";", ",":
            name = chr
            funcToken = immediates[chr]
            i++
            break loop
        }
        
        switch parseState {
        case "out":
            switch chr {
            case " ", "\t":
            default:
                parseState := "in"
                startToken := i
            }
        case "in":
            switch chr {
            case " ", "\t", "\n":
                name := s.Code[startToken:i]
                if i != "\n" {
                    i++
                }
                if builtin, ok := builtins["name"]; ok {
                    funcToken = func(s *State) State {
                        s.CurFunc = builtin
                        return s
                    }
                    break loop
                }
                if immediate, ok := builtins["name"]; ok {
                    funcToken = func(s *State) State {
                        return immediate(s)
                    }
                    break loop
                }

				if f, err := strconv.ParseFloat(name, 64); err == nil {
                    funcToken = func(s *State) State {
                        s.Push(f)
                        return s
                    }
                    break loop
                }

                if len(name) >= 1 && name[0] == "." {
                    funcToken = func(s *State) State {
                        s.Push(name[1:])
                        return s
                    }
                    break loop
                }

                _, v := s.findParentAndValue(name)
                switch v := v.(type) {
                case *Func:
                    funcToken = func(s *State) State {
                        v := s.Get(name)
                        s.CurFunc = v
                        return s
                    }
                default:
                    funcToken = func(s *State) State {
                        s.Push(Token{Name: name})
                        return s
                    }
                }
                break loop
            }
        }
    }

    if i == len(s.Code) {
        name = "done"
        funcToken = immediates[name]
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
func (state *State) GetVal(val any) any {
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


/*

if x is 3
    say "hello"
end


number a 36


record person
    name "Drew"
end

list numbers
    30 40 50
end






if ( a is 3 )

*/
	// FuncTokens         []func(*State) *State
	// FuncTokenStack     [][]func(*State) *State
	// FuncTokenSpots     []int // position of the first "argument" in vals, even tho it can grab from earlier
	// FuncTokenSpotStack [][]int
	// FuncTokenNames     []string
	// FuncTokenNameStack [][]string
 //
	// NewlineSpot              int
	// Breakables []Breakable

// type Func struct {
// 	FileName string
// 	I        int
// 	EndI     int
// 	// Note: the code and the cache should be bundled? (check perf)
// 	Code         string
// 	// ICache []*ICache
// 
// 	// TODO check these caches, or combine them ?
// 	// in a function are they correctly copied?
// 	Params            []string
// 	LexicalParent     *State
// 	Builtin           func(state *State) *State
// 	Name              string
// 
// 	// oneliner serves 2 things
// 	// one after def: func: loop: each: if:
// 	// and the other the state in the function
// 	// very closely related?
// 	OneLiner bool
// }
// 
// type Breakable struct {
//     Indent string
//     Type string
// }
