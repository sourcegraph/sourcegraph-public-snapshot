package rolebinding

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleRoleBinding() {
	rb, _ := NewRoleBinding("test", "sourcegraph")

	jrb, _ := json.Marshal(rb)
	fmt.Println(string(jrb))

	yrb, _ := yaml.Marshal(rb)
	fmt.Println(string(yrb))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"roleRef":{"apiGroup":"","kind":"","name":""}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// roleRef:
	//   apiGroup: ""
	//   kind: ""
	//   name: ""
}
