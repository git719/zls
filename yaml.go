// yaml.go

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

func LoadFileYAML(filePath string) (yamlObject map[interface{}]interface{}) {
	// Read/load/decode given filePath as some YAML object
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = yaml.Unmarshal(fileContent, &yamlObject)
	if err != nil {
		log.Println(err)
		return nil
	}
	return yamlObject
}

func PrintYAML(yamlObject interface{}) {
	pretty, err := yaml.Marshal(&yamlObject)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Printf(string(pretty))
	}
}
