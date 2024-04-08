package serviceaccount

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleServiceAccount() {
	sa, _ := NewServiceAccount("test", "sourcegraph")

	sat, _ := json.Marshal(sa)
	fmt.Println(string(sat))

	yat, _ := yaml.Marshal(sa)
	fmt.Println(string(yat))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
}
