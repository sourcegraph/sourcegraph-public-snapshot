package compute

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold/v2"
)

func Test_scanTemplate(t *testing.T) {
	test := func(input string) string {
		t := scanTemplate([]byte(input))
		return toJSONString(t)
	}

	autogold.Expect(`[{"constant":"artifcats: "},{"variable":"$repo"}]`).
		Equal(t, test("artifcats: $repo"))

	autogold.Expect(`[{"constant":"$"},{"variable":"$foo"},{"constant":" $"},{"variable":"$bar"}]`).
		Equal(t, test("$$foo $$bar"))

	autogold.Expect(`[{"variable":"$repo"},{"constant":"(derp)"}]`).
		Equal(t, test(`$repo(derp)`))

	autogold.Expect(`[{"variable":"$repo"},{"constant":":"},{"variable":"$file"},{"constant":" "},{"variable":"$content"}]`).
		Equal(t, test(`$repo:$file $content`))

	autogold.Expect(`[{"variable":"$repo"},{"variable":"$file"}]`).
		Equal(t, test("$repo$file"))

	autogold.Expect(`[{"constant":"$"},{"variable":"$foo"},{"variable":"$bar"},{"constant":"$"}]`).
		Equal(t, test("$$foo$bar$"))

	autogold.Expect(`[{"variable":"$bar"}]`).
		Equal(t, test("$bar"))

	autogold.Expect(`[{"variable":"$repo"},{"constant":" "}]`).
		Equal(t, test(`$repo\ `))

	autogold.Expect(`[{"constant":"$repo "}]`).
		Equal(t, test(`\$repo `))
}

func Test_templatize(t *testing.T) {
	autogold.Expect("artifcats: {{.Repo}}").
		Equal(t, templatize("artifcats: $repo", &MetaEnvironment{}))

	autogold.Expect("artifcats: {{.Repo}} $1").
		Equal(t, templatize("artifcats: $repo $1", &MetaEnvironment{}))
}

func Test_substituteMetaVariables(t *testing.T) {
	test := func(input string, env *MetaEnvironment) string {
		t, err := substituteMetaVariables(input, env)
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return t
	}

	autogold.Expect("artifcats: $1 $foo hi").
		Equal(t, test(
			"artifcats: $1 $foo $author",
			&MetaEnvironment{Author: "hi"},
		))
}
