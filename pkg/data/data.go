package data

import (
	"encoding/json"
	"os"

	autov1alpha1 "github.com/nokia/k8s-ipam/apis/auto/v1alpha1"
	invv1alpha1 "github.com/nokia/k8s-ipam/apis/inv/v1alpha1"
	"sigs.k8s.io/yaml"
)

func GetReplicaSet(fileName string) *autov1alpha1.ReplicaSet {
	b, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	//fmt.Println("replicaset raw data:\n", string(b))
	o := &autov1alpha1.ReplicaSet{}
	if err := yaml.Unmarshal(b, o); err != nil {
		panic(err)
	}
	return o
}

func GetEndpoint(fileName string) map[string]any {
	b, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	//fmt.Println("ep raw data:\n", string(b))

	ep := &invv1alpha1.Endpoint{}
	if err := yaml.Unmarshal(b, ep); err != nil {
		panic(err)
	}

	b, err = json.MarshalIndent(ep, "", "  ")
	if err != nil {
		panic(err)
	}
	x := map[string]any{}
	if err := json.Unmarshal(b, &x); err != nil {
		panic(err)
	}

	return x
}
