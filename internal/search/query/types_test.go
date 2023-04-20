package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoHasDescription(t *testing.T) {
	ps := Parameters{
		Parameter{
			Field:      FieldRepo,
			Value:      "has.description(test)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.description(test input)",
			Annotation: Annotation{Labels: IsPredicate},
		},
	}

	want := []string{
		"(?:test)",
		"(?:test).*?(?:input)",
	}

	require.Equal(t, want, ps.RepoHasDescription())
}

func TestRepoHasKVPs(t *testing.T) {
	ps := Parameters{
		Parameter{
			Field:      FieldRepo,
			Value:      "has(key:value)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.tag(tag)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.key(key)",
			Annotation: Annotation{Labels: IsPredicate},
		},
	}

	value := "value"
	want := []RepoKVPFilter{
		{Key: "key", Value: &value, Negated: false, KeyOnly: false},
		{Key: "tag", Value: nil, Negated: false, KeyOnly: false},
		{Key: "key", Value: nil, Negated: false, KeyOnly: true},
	}

	require.Equal(t, want, ps.RepoHasKVPs())
}

func TestRepoHasFileContent(t *testing.T) {
	ps := Parameters{
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(bare_path)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(path:path)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(path:path content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(-path:path)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(-content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(-path:path content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(path:path -content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.file(-path:path -content:content)",
			Annotation: Annotation{Labels: IsPredicate},
		},
	}

	want := []RepoHasFileContentArgs{
		{Path: "bare_path", Content: "", PathNegated: false, ContentNegated: false},
		{Path: "path", Content: "", PathNegated: false, ContentNegated: false},
		{Path: "", Content: "content", PathNegated: false, ContentNegated: false},
		{Path: "path", Content: "content", PathNegated: false, ContentNegated: false},
		{Path: "path", Content: "", PathNegated: true, ContentNegated: false},
		{Path: "", Content: "content", PathNegated: false, ContentNegated: true},
		{Path: "path", Content: "content", PathNegated: true, ContentNegated: false},
		{Path: "path", Content: "content", PathNegated: false, ContentNegated: true},
		{Path: "path", Content: "content", PathNegated: true, ContentNegated: true},
	}

	require.Equal(t, want, ps.RepoHasFileContent())
}
