// debug.go

package main

func TestFunction() {
	// url := "/providers/Microsoft.Management/getEntities"
	// url := "/providers/Microsoft.Management/managementGroups"
	// url := "/providers/Microsoft.Management/managementGroups/" + tenant_id + "/descendants"
	url := "/providers/Microsoft.Management/managementGroups/" + tenant_id
	params := map[string]string{
		"api-version": "2022-04-01",
		"$expand":     "children",
		"$recurse":    "true",
	}
	r := ApiGet(az_url+url, az_headers, params, false)
	ApiErrorCheck(r, trace())
	PrintJson(r)

	// Test MergeObjects function
	// filePath := filepath.Join("/Users/user1/obj.json")
	// objRaw := LoadFileJson(filePath)
	// if objRaw == nil {
	// 	die("Error loading file\n")
	// }
	// obj := objRaw.(map[string]interface{})

	// filePath2 = filepath.Join("/Users/user1/obj2.json")
	// objRaw2 = LoadFileJson(filePath2)
	// if objRaw2 == nil {
	// 	die("Error loading file\n")
	// }
	// obj2 := ojbRaw2.(map[string]interface{})

	// PrintJson(obj)
	// PrintJson(obj2)
	// x := MergeObjects(obj, obj2)
	// PrintJson(x)

	exit(0)
}
