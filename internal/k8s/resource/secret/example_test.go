package secret

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleSecret() {
	s, _ := NewSecret("test", "sourcegraph")

	js, _ := json.Marshal(s)
	fmt.Println(string(js))

	ys, _ := yaml.Marshal(s)
	fmt.Println(string(ys))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"immutable":false,"type":"Opaque"}
	// immutable: false
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// type: Opaque
}
