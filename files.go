// files.go

package main

import (
	"os"
	"path/filepath"
)

func RemoveFile(filePath string) {
	if FileExist(filePath) {
		if err := os.Remove(filePath); err != nil {
			panic(err.Error())
		}
	}
}

func FileUsable(filePath string) (e bool) {
	// True if file EXISTS && has SOME content && is less than 30 days old
	if FileExist(filePath) && FileSize(filePath) > 0 {
		return true
	}
	return false
}

func FileExist(filePath string) (e bool) {
	if _, err := os.Stat(filePath); err == nil || os.IsExist(err) {
		return true
	}
	return false
}

func FileNotExist(filePath string) (e bool) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	return false
}

func FileSize(filePath string) int64 {
	f, err := os.Stat(filePath)
	if err != nil { return 0 }
	return f.Size()
}

func FileModTime(filePath string) int {
	// Modified time in Unix epoch
	f, err := os.Stat(filePath)
	if err != nil { return 0 }
	return int(f.ModTime().Unix())
}

func RemoveCacheFile(t string) {
	// Remove cache file for objects of type t, or all of them
	filePath := ""
	switch t {
	case "t": // Token file is a little special: It doesn't use tenant ID
		filePath = filepath.Join(confdir, "accessTokens.json")
		RemoveFile(filePath)
	case "d", "a", "s", "u", "g", "sp", "ap", "ra", "rd":
		filePath = filepath.Join(confdir, tenant_id+"_"+oMap[t]+".json")
		RemoveFile(filePath)
		filePath = filepath.Join(confdir, tenant_id+"_"+oMap[t]+"_deltaLink.json")
		RemoveFile(filePath)
	case "all":
		for _, t := range oMap {
			filePath = filepath.Join(confdir, tenant_id+"_"+t+".json")
			RemoveFile(filePath)
			filePath = filepath.Join(confdir, tenant_id+"_"+t+"_deltaLink.json")
			RemoveFile(filePath)
		}
	}
	exit(0)
}

func CompareSpecfile(filePath string) {
	if FileNotExist(filePath) || FileSize(filePath) < 1 {
		die("File does not exist, or is zero size\n")
	}
    ft, t, x := GetObjectFromFile(filePath)
	if ft != "JSON" && ft != "YAML" {
        die("File is not in JSON nor YAML format\n")
    }
    if t != "d" && t != "a" {
        die("This " + ft + " file is not a role definition nor an assignment specfile\n")
    }
	
    print("==== SPECFILE ============================\n")
    PrintObject(t, x)
    print("==== AZURE ===============================\n")
    if t == "d" {
        y := GetAzRoleDefinition(x)
        if y == nil {
            print("Above definition does NOT exist in current Azure tenant\n")
        } else {
            PrintObject("d", y)
        }
    } else {
        y := GetAzRoleAssignment(x)
        if y == nil {
            print("Above assignment does NOT exist in current Azure tenant\n")
        } else {
            PrintObject("a", y)
        }
    }
    exit(0)	
}
