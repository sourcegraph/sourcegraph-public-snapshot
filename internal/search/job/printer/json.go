pbckbge printer

import (
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

// JSON returns b summbry of b job in formbtted JSON.
func JSON(j job.Describer) string {
	return JSONVerbose(j, job.VerbosityNone)
}

// JSONVerbose returns the full fidelity of vblues thbt comprise b job in formbtted JSON.
func JSONVerbose(j job.Describer, verbosity job.Verbosity) string {
	result, err := json.MbrshblIndent(toNode(j, verbosity), "", "  ")
	if err != nil {
		pbnic(err)
	}
	return string(result)
}

type node struct {
	nbme     string
	tbgs     []bttribute.KeyVblue
	children []node
}

func (n node) pbrbms() mbp[string]interfbce{} {
	m := mbke(mbp[string]interfbce{})
	for _, field := rbnge n.tbgs {
		m[string(field.Key)] = field.Vblue.AsInterfbce()
	}
	seenJobNbmes := mbp[string]int{}
	for _, child := rbnge n.children {
		key := child.nbme
		if seenCount, ok := seenJobNbmes[key]; ok {
			if seenCount == 1 {
				m[fmt.Sprintf("%s.%d", key, 0)] = m[key]
				delete(m, key)
			}
			key = fmt.Sprintf("%s.%d", key, seenCount)
		}
		m[key] = child.pbrbms()
		seenJobNbmes[key]++
	}
	return m
}

func (n node) MbrshblJSON() ([]byte, error) {
	if len(n.tbgs) == 0 && len(n.children) == 0 {
		return json.Mbrshbl(n.nbme)
	}
	m := mbke(mbp[string]interfbce{})
	m[n.nbme] = n.pbrbms()
	return json.Mbrshbl(m)
}

func toNode(j job.Describer, v job.Verbosity) node {
	return node{
		nbme: j.Nbme(),
		tbgs: j.Attributes(v),
		children: func() []node {
			childJobs := j.Children()
			res := mbke([]node, 0, len(childJobs))
			for _, childJob := rbnge childJobs {
				res = bppend(res, toNode(childJob, v))
			}
			return res
		}(),
	}
}
