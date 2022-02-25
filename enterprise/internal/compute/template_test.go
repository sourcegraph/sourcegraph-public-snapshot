package compute

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"
)

func Test_scanTemplate(t *testing.T) {
	test := func(input string) string {
		t := scanTemplate([]byte(input))
		return toJSONString(t)
	}

	autogold.Want(
		"basic template",
		`[{"constant":"artifcats: "},{"variable":"$repo"}]`).
		Equal(t, test("artifcats: $repo"))

	autogold.Want(
		"multiple $",
		`[{"constant":"$"},{"variable":"$foo"},{"constant":" $"},{"variable":"$bar"}]`).
		Equal(t, test("$$foo $$bar"))

	autogold.Want(
		"terminating variable",
		`[{"variable":"$repo"},{"constant":"(derp)"}]`).
		Equal(t, test(`$repo(derp)`))

	autogold.Want(
		"consecutive variables with separator",
		`[{"variable":"$repo"},{"constant":":"},{"variable":"$file"},{"constant":" "},{"variable":"$content"}]`).
		Equal(t, test(`$repo:$file $content`))

	autogold.Want(
		"consecutive variables no separator",
		`[{"variable":"$repo"},{"variable":"$file"}]`).
		Equal(t, test("$repo$file"))

	autogold.Want(
		"terminating variables with trailing $",
		`[{"constant":"$"},{"variable":"$foo"},{"variable":"$bar"},{"constant":"$"}]`).
		Equal(t, test("$$foo$bar$"))

	autogold.Want(
		"end-of-template variable",
		`[{"variable":"$bar"}]`).
		Equal(t, test("$bar"))

	autogold.Want(
		"space escaping",
		`[{"variable":"$repo"},{"constant":" "}]`).
		Equal(t, test(`$repo\ `))

	autogold.Want(
		"metachar escaping",
		`[{"constant":"$repo "}]`).
		Equal(t, test(`\$repo `))
}

func Test_templatize(t *testing.T) {
	autogold.Want(
		"basic templatize",
		"artifcats: {{.Repo}}").
		Equal(t, templatize("artifcats: $repo"))

	autogold.Want(
		"exclude regex var in templatize",
		"artifcats: {{.Repo}} $1").
		Equal(t, templatize("artifcats: $repo $1"))
}

func Test_substituteMetaVariables(t *testing.T) {
	test := func(input string, env *MetaEnvironment) string {
		t, err := substituteMetaVariables(input, env)
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return t
	}

	autogold.Want(
		"substitute for meta values in interface",
		"artifcats: $1 $foo hi").
		Equal(t, test(
			"artifcats: $1 $foo $author",
			&MetaEnvironment{Author: "hi"},
		))
}
