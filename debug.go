// debug.go

package main

func TestFunction() {
	// url := "/providers/Microsoft.Management/getEntities"
	// url := "/providers/Microsoft.Management/managementGroups"
	// url := "/providers/Microsoft.Management/managementGroups/" + tenant_id + "/descendants"
	url := "/providers/Microsoft.Management/managementGroups/" + tenant_id
	params := map[string]string{
		"api-version": "2020-05-01",
		"$expand":     "children",
		"$recurse":    "true",
	}
	r := APIGet(az_url+url, az_headers, params, false)
	PrintJSON(r)

	// Test MergeObjects function
	// f := filepath.Join("/Users/user1/obj.json")
	// l := LoadFileJSON(f)
	// if l == nil {
	// 	print("Error loading file\n")
	// 	exit(0)
	// }
	// obj := l.(map[string]interface{})

	// f = filepath.Join("/Users/user1/obj2.json")
	// l = LoadFileJSON(f)
	// if l == nil {
	// 	print("Error loading file\n")
	// 	exit(0)
	// }
	// obj2 := l.(map[string]interface{})

	// PrintJSON(obj)
	// PrintJSON(obj2)
	// x := MergeObjects(obj, obj2)
	// PrintJSON(x)

	exit(0)
}
