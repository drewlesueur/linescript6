package linescript6

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
	// "log"
)

// 1-based index List type backed by []any.
type List struct {
	TheSlice []any
}

// MarshalJSON implements custom JSON serialization for *List.
func (l *List) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.TheSlice)
}

// NewList returns a new empty *List.
func NewList() *List {
	return &List{TheSlice: make([]any, 0)}
}

// NewListWithCapacity returns a new *List with given initial size.
func NewListWithCapacity(size int) *List {
	return &List{TheSlice: make([]any, 0, size)}
}

func NewListFromStringSlice(slice []string) *List {
	newSlice := make([]any, len(slice))
	for i, v := range slice {
		newSlice[i] = v
	}
	return &List{TheSlice: newSlice}
}

// Push appends an element to the List.
func (l *List) Push(item any) {
	l.TheSlice = append(l.TheSlice, item)
}

// Pop removes and returns the last element. Returns nil if empty.
func (l *List) Pop() any {
	n := len(l.TheSlice)
	if n == 0 {
		return nil
	}
	val := l.TheSlice[n-1]
	l.TheSlice = l.TheSlice[:n-1]
	return val
}

// Shift removes and returns the first element. Returns nil if empty.
func (l *List) Shift() any {
	n := len(l.TheSlice)
	if n == 0 {
		return nil
	}
	val := l.TheSlice[0]
	l.TheSlice = l.TheSlice[1:]
	return val
}

// Unshift adds an element to the front of the List.
func (l *List) Unshift(item any) {
	l.TheSlice = append([]any{item}, l.TheSlice...)
}

// GetAt returns the value at a 1-based index. Returns nil if out of range.
func (l *List) Get(index int) any {
	if index == 0 {
		return nil
	}
	if index < 1 {
		return l.TheSlice[len(l.TheSlice)+index]
	}
	if index-1 >= len(l.TheSlice) {
		return nil
	}
	return l.TheSlice[index-1]
}

// SetAt sets the value at a 1-based index. Does nothing if out of range.
func (l *List) Set(index int, value any) {
	if index > len(l.TheSlice) {
		return
	}
	if index < 1 {
		index = len(l.TheSlice) + index + 1
	}
	l.TheSlice[index-1] = value
}

func (l *List) Length() int {
	return len(l.TheSlice)
}

// Join returns a string with each element converted to string (via toStringInternal)
// and joined by the given separator.
func (l *List) Join(sep string) string {
	items := make([]string, len(l.TheSlice))
	for i, v := range l.TheSlice {
		items[i] = toStringInternal(v)
	}
	return strings.Join(items, sep)
}

// Len returns the number of elements in the List.
func (l *List) Len() int {
	return len(l.TheSlice)
}

func (l *List) Slice(startInt, endInt int) *List {
	data := l.TheSlice
	if len(data) == 0 {
		return NewList()
	}
	n := len(data)
	if startInt < 0 {
		startInt = n + startInt + 1
	}
	if startInt <= 0 {
		startInt = 1
	}
	if startInt > n {
		return NewList()
	}
	if endInt < 0 {
		endInt = n + endInt + 1
	}
	if endInt <= 0 {
		return NewList()
	}
	if endInt > n {
		endInt = n
	}
	if startInt > endInt {
		return NewList()
	}
	sliced := make([]any, endInt-startInt+1)
	copy(sliced, data[startInt-1:endInt])
	return &List{TheSlice: sliced}
}

type RecordCache struct {
	Key   string
	Index int
	Value any
}

const cacheSize = 10

type Record struct {
	Values     []any
	KeyToIndex map[string]int
	Keys       []string
	Cache      []*RecordCache
}

// NewRecord returns a new empty Record.
func NewRecord() *Record {
	return &Record{
		Values:     make([]any, 0),
		KeyToIndex: make(map[string]int),
		Keys:       make([]string, 0),
		Cache:      make([]*RecordCache, cacheSize, cacheSize),
	}
}
func (r *Record) UpdateCache(key string, idx int, v any) {
	dataPtr := uintptr((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	slot := dataPtr % cacheSize
	rCache := r.Cache[slot]
	if rCache == nil {
		rCache = &RecordCache{}
		r.Cache[slot] = rCache
	}
	rCache.Value = v
	rCache.Key = key
	rCache.Index = idx

}

func (r *Record) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(r.Keys))
	for i, key := range r.Keys {
		m[key] = r.Values[i]
	}
	return json.Marshal(m)
}

func (r *Record) Length() int {
	return len(r.Keys)
}

// Set assigns value to key. If key is new, it appends it.
func (r *Record) Set(key string, value any) {
	if idx, ok := r.KeyToIndex[key]; ok {
		r.Values[idx] = value
		r.UpdateCache(key, idx, value)
	} else {
		r.Keys = append(r.Keys, key)
		r.Values = append(r.Values, value)
		r.KeyToIndex[key] = len(r.Values) - 1
		r.UpdateCache(key, len(r.Values)-1, value)
	}
}

// Get returns the value for key. The bool is false if key was not present.
func (r *Record) Get(key string) any {
	return nil
	if r == nil {
		return nil
	}

	dataPtr := uintptr((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	slot := dataPtr % cacheSize
	rCache := r.Cache[slot]
	if rCache != nil && rCache.Key == key {
		return rCache.Value
	}

	if idx, ok := r.KeyToIndex[key]; ok {
		return r.Values[idx]
	}
	return nil
}
func (r *Record) Has(key string) bool {
	if r == nil {
		return false
	}

	if _, ok := r.KeyToIndex[key]; ok {
		return true
	}
	return false
}

func (r *Record) GetHas(key string) (any, bool) {
	if r == nil {
		return nil, false
	}

	dataPtr := uintptr((*reflect.StringHeader)(unsafe.Pointer(&key)).Data)
	slot := dataPtr % cacheSize
	rCache := r.Cache[slot]
	if rCache != nil && rCache.Key == key {
		return rCache.Value, true
	}

	idx, ok := r.KeyToIndex[key]
	if ok {
		return r.Values[idx], true
	}
	return nil, false
}

// Delete removes the given key (and associated value) from the Record if it exists.
func (r *Record) Delete(key string) {
	idx, ok := r.KeyToIndex[key]
	if !ok {
		return
	}
	// Just clear the entry, don't reorder slices
	r.Keys[idx] = ""    // or some special marker; "" chosen for now
	r.Values[idx] = nil // clear value
	delete(r.KeyToIndex, key)
}

// GetIndex returns the value at the given index or nil if out of range.
func (r *Record) GetIndex(index int) any {
	if index < 0 || index >= len(r.Values) {
		return nil
	}
	return r.Values[index]
}

func toStringInternal(a any) string {
	switch a := a.(type) {
	case string:
		return a
	case map[string]any:
		return ToJsonF(a)
	case *Record:
		return ToJsonF(a)
	case *[]any, []any:
		return ToJsonF(a)
	case *List:
		return ToJsonF(a)
	case int:
		return strconv.Itoa(a)
	case int64:
		return strconv.Itoa(int(a))
	case float64:
		return strconv.FormatFloat(a, 'f', -1, 64)
	case bool:
		if a {
			return "true"
		}
		return "false"
	case nil:
		return "<nil>"
	case func(*Record) Record:
		// todo use unsafe ptr to see what it is
		return "a func, immediate"
	// case *Reader:
	// 	// TODO: read until a certain threshold,
	// 	// then use files?
	// 	// TODO: if this is part of "say" you can also consider just copying to stdout
	// 	// also execBash family should likely return a Reader.
	// 	b, err := io.ReadAll(a.Reader)
	// 	if err != nil {
	// 		panic(err) // ?
	// 	}
	// 	// "reset" the reader for later use
	// 	// part of the experiment to make readers and strings somewhat interchangeable
	// 	a.Reader = bytes.NewReader(b)
	//
	// 	return string(b)
	case uintptr:
		return fmt.Sprintf("%v", a)
	default:
		return fmt.Sprintf("toString: unknown type: type is %T, value is %#v\n", a, a)
	}
	return ""
}

// func toIntInternal(a any) int {
// 	switch a := a.(type) {
// 	case bool:
// 		if a {
// 			return 1
// 		}
// 		return 0
// 	case float64:
// 		return int(math.Floor(a))
// 	case int:
// 		return a
// 	case int64:
// 		return int(a)
// 	case string:
// 		if f, err := strconv.ParseFloat(a, 64); err == nil {
// 			return int(math.Floor(f))
// 		}
// 		return 0
// 	}
// 	return 0
// }
// func toInt(a any) any {
// 	return toIntInternal(a)
// }

// type String struct{
//     Value string
//     ID int
// }
