// users.go

package main

import (
	"fmt"
)

func PrintUser(x map[string]interface{}) {
	// Print user object in YAML-like style format
	if x["id"] == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "userPrincipalName", "mailNickname", "onPremisesSamAccountName",
		"onPremisesDomainName", "onPremisesUserPrincipalName"}
	for _, i := range list {
		fmt.Printf("%-28s %s\n", i+":", StrVal(x[i]))
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
	memberOf := GetObjectMemberOfs("u", id) // For this User object
	PrintMemberOfs("u", memberOf)
}
