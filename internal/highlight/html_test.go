package highlight

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"
)

func TestMultilineOccurrence(t *testing.T) {
	code := `namespace MyCompanyName.MyProjectName;

[DependsOn(
    // ABP Framework packages
    typeof(AbpAspNetCoreMvcModule)
)]`

	document := &scip.Document{
		Occurrences: []*scip.Occurrence{
			// Past end range occurrence
			{
				Range:      []int32{7, 0, 1},
				SyntaxKind: scip.SyntaxKind_BooleanLiteral,
			},
		},
	}

	rows := map[int32]bool{}
	scipToHTML(code, document, func(row int32) {
		rows[row] = true
	}, func(kind scip.SyntaxKind, line string) {}, nil)

	if len(rows) != 6 {
		t.Error("Should only add once per row, and should skip the 7th row (since it doesn't exist)")
	}
}

func TestMultilineOccurrence2(t *testing.T) {
	code := `namespace MyCompanyName.MyProjectName;

[DependsOn(
    // ABP Framework packages
    typeof(AbpAspNetCoreMvcModule)
)]`

	document := &scip.Document{
		Occurrences: []*scip.Occurrence{
			// Valid range at the beginning, should be skipped.
			{
				Range:      []int32{0, 0, 8},
				SyntaxKind: scip.SyntaxKind_IdentifierNamespace,
			},

			// Past end range occurrence, should not cause panic
			{
				Range:      []int32{2, 0, 30, 4},
				SyntaxKind: scip.SyntaxKind_BooleanLiteral,
			},
		},
	}

	// Should stay false, because we should skip this identifier
	// due to the valid lines that is passed to lsifToHTML
	sawNamespaceIdentifier := false

	rowsSeen := map[int32]bool{}
	scipToHTML(code, document, func(row int32) {
		rowsSeen[row] = true
	}, func(kind scip.SyntaxKind, line string) {
		if kind == scip.SyntaxKind_IdentifierNamespace {
			sawNamespaceIdentifier = true
		}

	}, map[int32]bool{
		0: false,
		1: false,
		2: true,
		3: false,
		4: false,
	})

	if len(rowsSeen) != 4 {
		t.Errorf("Should only add the rows from 2 until the end (due to weird multiline occurrence): %+v", rowsSeen)
	}

	if sawNamespaceIdentifier {
		t.Error("Should not have seen identifier for module, because line was skipped")
	}
}

//go:embed testdata/telemetry.scip
var telemetryDocument string

//go:embed testdata/telemetry-raw.txt
var telemetryText string

func getDocument(t *testing.T, payload string) *scip.Document {
	var document scip.Document
	err := proto.Unmarshal([]byte(payload), &document)
	if err != nil {
		t.Fatal(err)
	}

	return &document
}

func snapshotDocument(t *testing.T, document *scip.Document, code string, validLines map[int32]bool) []string {
	result := ""
	addRow := func(row int32) {
		if result != "" {
			result += "\n"
		}

		result += fmt.Sprintf("%03d: ", row)
	}

	addText := func(kind scip.SyntaxKind, line string) {
		if line == "" {
			t.Fatalf("Line should not be empty: %s\n%s", kind, result)
		}

		line = strings.ReplaceAll(line, "\n", "\\n")

		if kind == scip.SyntaxKind_UnspecifiedSyntaxKind {
			result += line
		} else {
			result += fmt.Sprintf("<%s>%s</>", kind, line)
		}
	}

	scipToHTML(code, document, addRow, addText, validLines)
	return strings.Split(result, "\n")
}

func TestMultilineTypescriptCommentEntire(t *testing.T) {
	document := getDocument(t, telemetryDocument)
	splitResult := snapshotDocument(t, document, telemetryText, nil)

	autogold.Expect([]string{
		"000: <Keyword>import</> * <Keyword>as</> <Identifier>sourcegraph</> <Keyword>from</> <StringLiteral>'</><StringLiteral>./api</><StringLiteral>'</>",
		"001: \\n",
		"002: <Comment>/**</>",
		"003: <Comment> * A wrapper around telemetry events. A new instance of this class</>",
		"004: <Comment> * should be instantiated at the start of each action as it handles</>",
		"005: <Comment> * latency tracking.</>",
		"006: <Comment> */</>",
		"007: <Keyword>export</> <Keyword>class</> <IdentifierType>TelemetryEmitter</> {",
		"008:     <Keyword>private</> <Identifier>languageID</>: <IdentifierBuiltinType>string</>",
		"009:     <Keyword>private</> <Identifier>repoID</>: <IdentifierBuiltinType>number</>",
		"010:     <Keyword>private</> <Identifier>started</>: <IdentifierBuiltinType>number</>",
		"011:     <Keyword>private</> <Identifier>enabled</>: <IdentifierBuiltinType>boolean</>",
		"012:     <Keyword>private</> <Identifier>emitted</> = <Keyword>new</> <Identifier>Set</><<IdentifierBuiltinType>string</>>()",
		"013: \\n",
		"014:     <Comment>/**</>",
		"015: <Comment>     * Creates a new telemetry emitter object for a given</>",
		"016: <Comment>     * language ID and repository ID.</>",
		"017: <Comment>     * Emitting is enabled by default</>",
		"018: <Comment>     *</>",
		"019: <Comment>     * @param languageID The language identifier e.g. 'java'.</>",
		"020: <Comment>     * @param repoID numeric repository identifier.</>",
		"021: <Comment>     * @param enabled Whether telemetry is enabled.</>",
		"022: <Comment>     */</>",
		"023:     <IdentifierFunction>constructor</>(<Identifier>languageID</>: <IdentifierBuiltinType>string</>, <Identifier>repoID</>: <IdentifierBuiltinType>number</>, <Identifier>enabled</> = <IdentifierBuiltin>true</>) {",
		"024:         <IdentifierBuiltin>this</>.<Identifier>languageID</> = <Identifier>languageID</>",
		"025:         <IdentifierBuiltin>this</>.<Identifier>started</> = <Identifier>Date</>.<IdentifierFunction>now</>()",
		"026:         <IdentifierBuiltin>this</>.<Identifier>repoID</> = <Identifier>repoID</>",
		"027:         <IdentifierBuiltin>this</>.<Identifier>enabled</> = <Identifier>enabled</>",
		"028:     }",
		"029: \\n",
		"030:     <Comment>/**</>",
		"031: <Comment>     * Emit a telemetry event with a durationMs attribute only if the</>",
		"032: <Comment>     * same action has not yet emitted for this instance. This method</>",
		"033: <Comment>     * returns true if an event was emitted and false otherwise.</>",
		"034: <Comment>     */</>",
		"035:     <Keyword>public</> <IdentifierFunction>emitOnce</>(<Identifier>action</>: <IdentifierBuiltinType>string</>, <Identifier>args</>: <IdentifierBuiltinType>object</> = {}): <IdentifierBuiltinType>boolean</> {",
		"036:         <Keyword>if</> (<IdentifierBuiltin>this</>.<Identifier>emitted</>.<IdentifierFunction>has</>(<Identifier>action</>)) {",
		"037:             <Keyword>return</> <IdentifierBuiltin>false</>",
		"038:         }",
		"039: \\n",
		"040:         <IdentifierBuiltin>this</>.<Identifier>emitted</>.<IdentifierFunction>add</>(<Identifier>action</>)",
		"041:         <IdentifierBuiltin>this</>.<IdentifierFunction>emit</>(<Identifier>action</>, <Identifier>args</>)",
		"042:         <Keyword>return</> <IdentifierBuiltin>true</>",
		"043:     }",
		"044: \\n",
		"045:     <Comment>/**</>",
		"046: <Comment>     * Emit a telemetry event with durationMs and languageId attributes.</>",
		"047: <Comment>     */</>",
		"048:     <Keyword>public</> <IdentifierFunction>emit</>(<Identifier>action</>: <IdentifierBuiltinType>string</>, <Identifier>args</>: <IdentifierBuiltinType>object</> = {}): <Keyword>void</> {",
		"049:         <Keyword>if</> (!<IdentifierBuiltin>this</>.<Identifier>enabled</>) {",
		"050:             <Keyword>return</>",
		"051:         }",
		"052: \\n",
		"053:         <Keyword>try</> {",
		"054:             <Identifier>sourcegraph</>.<IdentifierFunction>logTelemetryEvent</>(<StringLiteral>`codeintel.</><StringLiteralEscape>${</><Identifier>action</><StringLiteralEscape>}</><StringLiteral>`</>, {",
		"055:                 ...<Identifier>args</>,",
		"056:                 <IdentifierAttribute>durationMs</>: <IdentifierBuiltin>this</>.<IdentifierFunction>elapsed</>(),",
		"057:                 <IdentifierAttribute>languageId</>: <IdentifierBuiltin>this</>.<Identifier>languageID</>,",
		"058:                 <IdentifierAttribute>repositoryId</>: <IdentifierBuiltin>this</>.<Identifier>repoID</>,",
		"059:             })",
		"060:         } <Keyword>catch</> {",
		"061:             <Comment>// Older version of Sourcegraph may have not registered this</>",
		"062:             <Comment>// command, causing the promise to reject. We can safely ignore</>",
		"063:             <Comment>// this condition.</>",
		"064:         }",
		"065:     }",
		"066: \\n",
		"067:     <Keyword>private</> <IdentifierFunction>elapsed</>(): <IdentifierBuiltinType>number</> {",
		"068:         <Keyword>return</> <Identifier>Date</>.<IdentifierFunction>now</>() - <IdentifierBuiltin>this</>.<Identifier>started</>",
		"069:     }",
		"070: }",
		"071: \\n",
	}).Equal(t, splitResult)
}

func TestMultilineTypescriptCommentRanges(t *testing.T) {
	document := getDocument(t, telemetryDocument)
	splitResult := snapshotDocument(t, document, telemetryText, map[int32]bool{
		6: true,
		7: true,
		8: true,
	})

	autogold.Expect([]string{
		"006: <Comment> */</>",
		"007: <Keyword>export</> <Keyword>class</> <IdentifierType>TelemetryEmitter</> {",
		"008:     <Keyword>private</> <Identifier>languageID</>: <IdentifierBuiltinType>string</>",
	}).Equal(t, splitResult)
}
