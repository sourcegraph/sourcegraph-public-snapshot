package deployment

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleDeployment() {
	d, _ := NewDeployment("test", "sourcegraph", "1.2.3")

	jd, _ := json.Marshal(d)
	fmt.Println(string(jd))

	yd, _ := yaml.Marshal(d)
	fmt.Println(string(yd))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"app.kubernetes.io/component":"test","app.kubernetes.io/name":"sourcegraph","app.kubernetes.io/version":"1.2.3","deploy":"sourcegraph"}},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"template":{"metadata":{"creationTimestamp":null},"spec":{"containers":null}},"strategy":{"type":"Recreate"},"minReadySeconds":10,"revisionHistoryLimit":10},"status":{}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     app.kubernetes.io/component: test
	//     app.kubernetes.io/name: sourcegraph
	//     app.kubernetes.io/version: 1.2.3
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec:
	//   minReadySeconds: 10
	//   replicas: 1
	//   revisionHistoryLimit: 10
	//   selector:
	//     matchLabels:
	//       app: test
	//   strategy:
	//     type: Recreate
	//   template:
	//     metadata:
	//       creationTimestamp: null
	//     spec:
	//       containers: null
	// status: {}
}
