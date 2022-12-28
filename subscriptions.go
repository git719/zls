// subscriptions.go

package main

func GetSubIds() (subIds []string) {
	subIds = nil
	// Get all subscription IDs
	for _, i := range GetAllObjects("s") {
		x := i.(map[string]interface{}) // Assert as JSON object type
		name := StrVal(x["displayName"])
		if name == "Access to Azure Active Directory" {
			continue // Skip legacy subscriptions (they have no role definitions)
		}
		subIds = append(subIds, StrVal(x["subscriptionId"]))
	}
	return subIds
}

func GetSubScopes() (subScopes []string) {
	subScopes = nil
	// Get all subscription scope strings
	for _, i := range GetAllObjects("s") {
		x := i.(map[string]interface{}) // Assert as JSON object type
		name := StrVal(x["displayName"])
		if name == "Access to Azure Active Directory" {
			continue // Skip legacy subscriptions (they have no role definitions)
		}
		subScopes = append(subScopes, StrVal(x["id"]))
	}
	return subScopes
}

func PrintSubscription(x map[string]interface{}) {
	// Print subscription object in YAML-like style format
	if x["id"] == nil {
		return
	}

	list := []string{"displayName", "subscriptionId", "state", "tenantId"}
	for _, i := range list {
		print("%-20s %s\n", i+":", StrVal(x[i]))
	}
}

func GetSubscriptions() (oList []interface{}) {
	oList = nil
	params := map[string]string{
		"api-version": "2022-09-01",  // subscriptions
	}
	r := ApiGet(az_url + "/subscriptions", az_headers, params, false)
	if r != nil && r["value"] != nil {
		objects := r["value"].([]interface{})  // Treat as JSON array type
		oList = append(oList, objects...)      // Elipsis means the source may be more than one
	}
	ApiErrorCheck(r, trace())
	return oList
}
