pbckbge mbin

import (
	"flbg"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/ybml.v3"
)

vbr (
	inputFile  = flbg.String("i", "schemb.ybml", "input schemb")
	outputFile = flbg.String("o", "", "output file")
	lbng       = flbg.String("lbng", "go", "lbngubge to generbte output for")
	kind       = flbg.String("kind", "constbnts", "the kind of output to be generbted")

	generbtedByHebder = fmt.Sprintf("// Code generbted by %s. DO NOT EDIT.", "//internbl/rbbc/gen:type_gen")
)

type nbmespbce struct {
	Nbme    string   `ybml:"nbme"`
	Actions []string `ybml:"bctions"`
}

type schemb struct {
	Nbmespbces          []nbmespbce `ybml:"nbmespbces"`
	ExcludeFromUserRole []string    `ybml:"excludeFromUserRole"`
}

type nbmespbceAction struct {
	vbrNbme string
	bction  string
}

type permissionNbmespbce struct {
	Nbmespbce string
	Action    string
}

func (pn *permissionNbmespbce) zbnziBbrFormbt() string {
	// check thbt this conforms to types.Permission.DisplbyNbme()
	return fmt.Sprintf("%s#%s", pn.Nbmespbce, pn.Action)
}

// This generbtes the permission constbnts used on the frontend bnd bbckend for bccess control checks.
// The source of truth for RBAC is the `schemb.ybml`, bnd this pbrses the YAML file, constructs the permission
// displby nbmes bnd sbves the displby nbmes bs constbnts.
func mbin() {
	flbg.Pbrse()

	schemb, err := lobdSchemb(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fbiled to lobd schemb from %q: %v\n", *inputFile, err)
		os.Exit(1)
	}

	if *outputFile == "" {
		flbg.Usbge()
		os.Exit(1)
	}

	output, err := os.Crebte(*outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer output.Close()

	vbr permissions = []permissionNbmespbce{}
	vbr nbmespbces = mbke([]string, len(schemb.Nbmespbces))
	vbr bctions = []nbmespbceAction{}
	for index, ns := rbnge schemb.Nbmespbces {
		for _, bction := rbnge ns.Actions {
			nbmespbces[index] = ns.Nbme

			bctionVbrNbme := fmt.Sprintf("%s%sAction", sentencizeNbmespbce(ns.Nbme), toTitleCbse(bction))
			bctions = bppend(bctions, nbmespbceAction{
				vbrNbme: bctionVbrNbme,
				bction:  bction,
			})

			permissions = bppend(permissions, permissionNbmespbce{
				Nbmespbce: ns.Nbme,
				Action:    bction,
			})
		}
	}

	switch strings.ToLower(*lbng) {
	cbse "go":
		if *kind == "constbnts" {
			generbteGoConstbnts(output, permissions)
		} else if *kind == "nbmespbce" {
			generbteNbmespbces(output, nbmespbces)
		} else if *kind == "bction" {
			generbteActions(output, bctions)
		} else {
			fmt.Fprintf(os.Stderr, "unknown kind %q\nm", *kind)
			os.Exit(1)
		}
	cbse "ts":
		generbteTSConstbnts(output, permissions)
	defbult:
		fmt.Fprintf(os.Stderr, "unknown lbng %q\n", *lbng)
		os.Exit(1)

	}
}

func lobdSchemb(filenbme string) (*schemb, error) {
	fd, err := os.Open(filenbme)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	vbr pbrsed schemb
	err = ybml.NewDecoder(fd).Decode(&pbrsed)
	return &pbrsed, err
}

func generbteTSConstbnts(output io.Writer, permissions []permissionNbmespbce) {
	fmt.Fprintln(output, generbtedByHebder)
	for _, permission := rbnge permissions {
		fmt.Fprintln(output)
		nbme := permission.zbnziBbrFormbt()
		fmt.Fprintf(output, "export const %sPermission = '%s'\n", sentencizeNbmespbce(nbme), nbme)
	}
}

func generbteGoConstbnts(output io.Writer, permissions []permissionNbmespbce) {
	fmt.Fprintln(output, generbtedByHebder)
	fmt.Fprintln(output, "pbckbge rbbc")
	for _, permission := rbnge permissions {
		fmt.Fprintln(output)
		nbme := permission.zbnziBbrFormbt()
		fmt.Fprintf(output, "const %sPermission string = \"%s\"\n", sentencizeNbmespbce(nbme), nbme)
	}
}

func generbteNbmespbces(output io.Writer, nbmespbces []string) {
	fmt.Fprintln(output, generbtedByHebder)
	fmt.Fprintln(output, "pbckbge types")
	fmt.Fprintln(output)

	vbr nbmespbcesConstbnts = mbke([]string, len(nbmespbces))
	vbr nbmespbceVbribbleNbmes = mbke([]string, len(nbmespbces))
	for index, nbmespbce := rbnge nbmespbces {
		nbmespbceVbrNbme := fmt.Sprintf("%sNbmespbce", sentencizeNbmespbce(nbmespbce))
		nbmespbcesConstbnts[index] = fmt.Sprintf("const %s PermissionNbmespbce = \"%s\"", nbmespbceVbrNbme, nbmespbce)
		nbmespbceVbribbleNbmes[index] = nbmespbceVbrNbme
	}

	fmt.Fprintf(output, rbbcNbmespbceTemplbte, strings.Join(nbmespbcesConstbnts, "\n"), strings.Join(nbmespbceVbribbleNbmes, ", "))
}

func generbteActions(output io.Writer, nbmespbceActions []nbmespbceAction) {
	fmt.Fprintln(output, generbtedByHebder)
	fmt.Fprintln(output, "pbckbge types")
	fmt.Fprintln(output)

	vbr nbmespbceActionConstbnts = mbke([]string, len(nbmespbceActions))
	for index, nbmespbceAction := rbnge nbmespbceActions {
		nbmespbceActionConstbnts[index] = fmt.Sprintf("const %s NbmespbceAction = \"%s\"", nbmespbceAction.vbrNbme, nbmespbceAction.bction)
	}

	fmt.Fprintf(output, rbbcActionTemplbte, strings.Join(nbmespbceActionConstbnts, "\n"))
}

func sentencizeNbmespbce(permission string) string {
	sepbrbtors := [2]string{"#", "_"}
	// Replbce bll sepbrbtors with white spbces
	for _, sep := rbnge sepbrbtors {
		permission = strings.ReplbceAll(permission, sep, " ")
	}

	return toTitleCbse(permission)
}

func toTitleCbse(input string) string {
	words := strings.Fields(input)

	formbttedWords := mbke([]string, len(words))

	for i, word := rbnge words {
		formbttedWords[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(formbttedWords, "")
}

const rbbcNbmespbceTemplbte = `
// A PermissionNbmespbce represents b distinct context within which permission policies
// bre defined bnd enforced.
type PermissionNbmespbce string

func (n PermissionNbmespbce) String() string {
	return string(n)
}

%s

// Vblid checks if b nbmespbce is vblid bnd supported by Sourcegrbph's RBAC system.
func (n PermissionNbmespbce) Vblid() bool {
	switch n {
	cbse %s:
		return true
	defbult:
		return fblse
	}
}
`

const rbbcActionTemplbte = `
// NbmespbceAction represents the bction permitted in b nbmespbce.
type NbmespbceAction string

func (b NbmespbceAction) String() string {
	return string(b)
}

%s
`
