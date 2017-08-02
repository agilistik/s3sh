package main

import (
	"strings"
)

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

func BuildPath (startPath, addPath string) []string  {
	startArray	:= strings.Split(startPath, "/")
	addArray	:= strings.Split(addPath, "/")
	if strings.Index(addPath, "/") == 0 {
		return addArray
	}
	resArray	:= make([]string, len(startArray) + len(addArray))
	copy (resArray, startArray)
	addIndex := len(startArray)
	
	for _, j := range (addArray) {
		if j == ".." {
			if addIndex > 1 {	
				addIndex = addIndex - 1
			}
		} else {
			if j != "." {
				resArray[addIndex] = j
				addIndex = addIndex + 1
			}
		}
	}
	return resArray[:addIndex]
	
}


