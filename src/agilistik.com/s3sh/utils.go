package main

type StrSet struct {
        set map[string]bool
}

func NewStrSet () *StrSet {
        return &StrSet{make(map[string]bool)}
}

func (set *StrSet) Add (s string) bool {
        _, found := set.set[s]
        set.set[s] = true
        return !found
}

func StringInSlice(a string, list [] string) bool {
        for _, b := range list {
                if b == a {
                        return true
                }
        }
        return false
}


