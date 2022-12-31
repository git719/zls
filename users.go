// users.go

package main

func PrintUser(x map[string]interface{}) {
	// Print user object in YAML format
	if x["id"] == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "userPrincipalName", "mailNickname", "onPremisesSamAccountName",
		"onPremisesDomainName", "onPremisesUserPrincipalName"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { print("%-28s %s\n", i+":", v) } // Only print non-null attributes
	}

	if x["otherMails"] != nil {
		otherMails := x["otherMails"].([]interface{})
		if len(otherMails) > 0 {
			print("otherMails:\n")
			for _, i := range otherMails {
				email := i.(string)
				print("  %s\n", email)
			}
		} else {
			print("%-28s %s\n", "otherMails:", "None")
		}
	}

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("u", id) // For this User object
	PrintMemberOfs("u", memberOf)
}
