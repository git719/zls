// strings.go

package main

import "strings"

func SubString(large, small string) bool {
	// Case insensitive substring search
	if strings.Contains(strings.ToLower(large), strings.ToLower(small)) {
		return true
	}
	return false
}

func LastElem(s, splitter string) string {
	split := strings.Split(s, splitter) // Split the string
	return split[len(split)-1]          // Return last element
}

func StrVal(x interface{}) string {
	// Return the best printable string value for given x variable
	if x != nil {
		switch sprint("%T", x) {
		case "bool":
			return sprint("%t", x)
		case "string":
			return x.(string)
		default:
			return "" // Blank for other types
		}
	}
	return "" // Blank if nil
}

func ItemInList(arg string, argList []string) bool {
	for _, value := range argList {
		if value == arg { return true }
	}
	return false
}

func PadSpaces(n int) {
	for i := 0; i < n; i++ {
		print(" ")
	}
}
