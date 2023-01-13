// debug.go

package main

import (
	"fmt"
	"os"
	//"runtime"
	//"path/filepath"
	// "github.com/git719/aza"
	// "github.com/git719/utl"
)
var (
	// Some basic colors
	Red = "\033[1;31m" ; Gre = "\033[0;32m" ; Blu = "\033[1;34m"
	Yel = "\033[0;33m" ; Pur = "\033[1;35m" ; Cya = "\033[0;36m"
    Rst = "\033[0m"
)

// func init() {
// 	if runtime.GOOS == "windows" {
// 		Red = "" ; Gre = "" ; Blu = "" ; Yel = "" ; Pur = "" ; Cya = "" ; Rst = ""
// 	}
// }

func TestFunction() {
	fmt.Println(Red + "Hellow world!" + Rst)
	fmt.Println(Gre + "Hellow world!" + Rst)
	fmt.Println(Blu + "Hellow world!" + Rst)
	fmt.Println(Yel + "Hellow world!" + Rst)
	fmt.Println(Pur + "Hellow world!" + Rst)
	fmt.Println(Cya + "Hellow world!" + Rst)

	// 	var z aza.AzaBundle
// 	var oMap MapString

	// cacheFile := "/Users/user1/.zls/3f550b9f-29b0-4ba6-ad61-c95f63104213_users.json"
	// deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	// fmt.Println(cacheFile)
	// fmt.Println(deltaLinkFile)

// 	// Set up the bundle of variables
// 	z, oMap = SetupVariables(&z, &oMap)
// 	fmt.Println(z.ConfDir)
// 	fmt.Println(z.CredsFile)
// 	fmt.Println(z.TokenFile)
// 	fmt.Println(oMap["d"])
// 	fmt.Println(oMap["m"])
// 	fmt.Println(oMap["sp"])
	
// 	// Setup the API tokens
// 	z = aza.SetupApiTokens(&z)

// 	// Get subscription with a specific filter
// 	subs := GetSubscriptions("as01", z)
// 	fmt.Println(len(subs))
// 	utl.PrintJson(subs)
	os.Exit(0)
}
