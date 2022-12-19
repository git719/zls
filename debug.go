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
	PrintJson(r)

	// Test MergeObjects function
	// f := filepath.Join("/Users/user1/obj.json")
	// l := LoadFileJson(f)
	// if l == nil {
	// 	die("Error loading file\n")
	// }
	// obj := l.(map[string]interface{})

	// f = filepath.Join("/Users/user1/obj2.json")
	// l = LoadFileJson(f)
	// if l == nil {
	// 	die("Error loading file\n")
	// }
	// obj2 := l.(map[string]interface{})

	// PrintJson(obj)
	// PrintJson(obj2)
	// x := MergeObjects(obj, obj2)
	// PrintJson(x)

	exit(0)
}
