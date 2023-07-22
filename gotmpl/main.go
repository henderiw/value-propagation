package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/henderiw/value-propagation/pkg/data"
	//"sigs.k8s.io/yaml"
)

const (
	index              = "index"
	key                = "key"
	value              = "value"
	var1               = "var1"
	replicaSetFileName = "../data/node-replicaset-gotempl.yaml"
	endpointFileName   = "../data/endpoint_leaf1-e1-1.yaml"
)

func main() {
	cr := data.GetReplicaSet(replicaSetFileName)
	fmt.Println("template data:\n", string(cr.Spec.Template.Raw))

	if cr.Spec.Replicas != nil {
		for i := 0; i < int(*cr.Spec.Replicas); i++ {
			result := new(bytes.Buffer)
			// TODO: add template custom functions
			tpl, err := template.New("default").Option("missingkey=zero").Parse(string(cr.Spec.Template.Raw))
			if err != nil {
				panic(err)
			}

			input := map[string]any{
				"var1":  data.GetEndpoint(endpointFileName),
				"index": i,
			}
			//fmt.Printf("var1:\n%s\n", input["var1"])

			err = tpl.Execute(result, input)
			if err != nil {
				panic(err)
			}

			var x map[string]any
			if err := json.Unmarshal(result.Bytes(), &x); err != nil {
				panic(err)
			}

			b, err := json.MarshalIndent(x, "", "  ")
			if err != nil {
				panic(err)
			}

			fmt.Printf("result %d:\n%v\n", i, string(b))

		}
	}
}
