// mgmtgroups.go

package main

import "fmt"

func MGType(typeIn string) string {
	switch typeIn {
	case "Microsoft.Management/managementGroups":
		return "ManagementGroup"
	case "Microsoft.Management/managementGroups/subscriptions", "/subscriptions":
		return "Subscription"
	default:
		return "??"
	}
}

func PrintManagementGroup(x map[string]interface{}) {
	// Print subscription object in YAML-like style format
	if x["id"] == nil {
		return
	}
	id := StrVal(x["name"])
	xProp := x["properties"].(map[string]interface{})
	Name := StrVal(xProp["displayName"])
	Type := MGType(StrVal(x["type"]))
	fmt.Printf("%-20s %s\n", "displayName:", Name)
	fmt.Printf("%-20s %s\n", "id:", id)
	fmt.Printf("%-20s %s\n", "type:", Type)
}

func GetManagementGroups() (oList []interface{}) {
	oList = nil
	url := az_url + "/providers/Microsoft.Management/" + oMap["m"]
	params := map[string]string{"api-version": "2020-05-01"}
	r := APIGet(url, az_headers, params, false)
	if r["value"] != nil {
		objects := r["value"].([]interface{}) // Treat as JSON array type
		oList = append(oList, objects...)     // Elipsis means the source may be more than one
	}
	return oList
}

func PrintMGChildren(indent int, children []interface{}) {
	for _, i := range children {
		child := i.(map[string]interface{})
		name := StrVal(child["displayName"])
		id := StrVal(child["name"])
		Type := MGType(StrVal(child["type"]))

		if name == "Access to Azure Active Directory" && Type == "Subscription" {
			continue // Skip legacy subscriptions. We don't care
		}

		PadSpaces(indent)
		padding := 38 - indent
		if padding < 12 {
			padding = 12
		}
		print("%-*s  %-38s  %s\n", padding, name, id, Type)
		if child["children"] != nil {
			descendants := child["children"].([]interface{})
			PrintMGChildren(indent+4, descendants) // Recurse to print additional children
		}
	}
}

func PrintManagementGroupTree() {
	// Get the entire MG and subscription hierarchy tree for the tenant
	url := "/providers/Microsoft.Management/managementGroups/" + tenant_id
	params := map[string]string{
		"api-version": "2020-05-01",
		"$expand":     "children",
		"$recurse":    "true",
	}
	r := APIGet(az_url+url, az_headers, params, false)
	if r["properties"] != nil {
		// Print everything under the hierarchy
		props := r["properties"].(map[string]interface{})
		name := StrVal(props["displayName"])
		id := StrVal(props["tenantId"])
		print("%-38s  %-38s  TENANT\n", name, id)
		if props["children"] != nil {
			children := props["children"].([]interface{})
			PrintMGChildren(4, children)
		}
	}
}
