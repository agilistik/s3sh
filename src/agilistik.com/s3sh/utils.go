package main

import (
	"strings"
)
/*
	 StrSet, NewStrSet, and Add  represent a set of strings.
	 Currently, only 'Add' function is necessary.
*/
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

/*
	Tracking history of commands
*/
type Hist struct {
	history [] string
	nextPos int
}

func NewHist (size int) *Hist {
	return &Hist{make([]string, size), 0}
}

func (hist *Hist) Add (s string) {
	hist.history[hist.nextPos] = s
	hist.nextPos = hist.nextPos + 1
	if hist.nextPos == len(hist.history)  {
		hist.nextPos = 0		
	}
}

/*
	Input:  starting piont, usually pwd; and, new path.
	Output:  the new path, if it is absolute.
		Otherwise, starting piont + new path, with correct resolution of '.' and '..'
	Output is an array representing the path.  It's caller's job to join it to build the path.
	This way, the caller can use the correct  path separator if necessary.
	Also, this provides  the ability to get any path element  (basedir, basepath, etc) at virtually no additioanl cost.
*/
func BuildPath (startPath, addPath string) []string  {
	startArray	:= strings.Split(startPath, "/")
	addArray	:= strings.Split(addPath, "/")
	if strings.Index(addPath, "/") == 0 {
		return addArray
	}
	resArray	:= make([]string, len(startArray) + len(addArray))
	copy (resArray, startArray)
// Remove empty strings from the beginning of resArray:
/*
	for i, j := range(resArray){
		if j != "" {
			resArray = resArray[i:]
			break
			
		}
	}	
*/
	addIndex := len(startArray)
	
	for _, j := range (addArray) {
		if j == ".." {
			if addIndex > 1 {	
				addIndex = addIndex - 1
			}
		} else {
			if j != "." && j != "" { 
				resArray[addIndex] = j
				addIndex = addIndex + 1
			}
		}
	}
	return resArray[:addIndex]
	
}

/*
	Input:  path
	Output:  bucket, prefix
*/
func BucketPrefix (path string) (bucket, prefix string) {
	pathArr := BuildPath("/", path)
	bucket = pathArr[1]
	prefix = strings.Join(pathArr[2:], "/") + "/"
	return bucket, prefix	
}

