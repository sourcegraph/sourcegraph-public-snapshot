package plan

import droneyaml "github.com/drone/drone-exec/yaml"

// buildLogMsg generates an empty CI test plan that prints msg to the
// build log.
func buildLogMsg(title, msg string) droneyaml.BuildItem {
	return droneyaml.BuildItem{
		Key: "Warning: " + title,
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image:       "library/alpine:3.2",
				Environment: droneyaml.MapEqualSlice([]string{"MSG=" + msg}),
			},
			Commands: []string{`echo "$MSG"`},
		},
	}
}
