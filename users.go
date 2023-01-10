// users.go

package main

import (
	"fmt"
	"github.com/git719/aza"
)

func PrintUser(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print user object in YAML format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "userPrincipalName", "mailNickname", "onPremisesSamAccountName",
		"onPremisesDomainName", "onPremisesUserPrincipalName"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { // Only print non-null attributes
			fmt.Printf("%-28s %s\n", i+":", v)
		}
	}

	if x["otherMails"] != nil {
		otherMails := x["otherMails"].([]interface{})
		if len(otherMails) > 0 {
			fmt.Printf("otherMails:\n")
			for _, i := range otherMails {
				email := i.(string)
				fmt.Printf("  %s\n", email)
			}
		} else {
			fmt.Printf("%-28s %s\n", "otherMails:", "None")
		}
	}

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("u", id, z, oMap) // For this User object
	PrintMemberOfs("u", memberOf)
}

