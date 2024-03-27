package role

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleRole() {
	r, _ := NewRole("test", "sourcegraph")

	jr, _ := json.Marshal(r)
	fmt.Println(string(jr))

	yr, _ := yaml.Marshal(r)
	fmt.Println(string(yr))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"rules":null}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// rules: null
}
