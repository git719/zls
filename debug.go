// debug.go

package main

import (
	"fmt"
	"os"
	"path/filepath"
	// "github.com/git719/aza"
	// "github.com/git719/utl"
)

func TestFunction() {
	fmt.Println("Hellow world!")
// 	var z aza.AzaBundle
// 	var oMap MapString

	cacheFile := "/Users/user1/.zls/3f550b9f-29b0-4ba6-ad61-c95f63104213_users.json"
	deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	fmt.Println(cacheFile)
	fmt.Println(deltaLinkFile)

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
