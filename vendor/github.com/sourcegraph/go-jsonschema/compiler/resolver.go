package compiler

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

func resolveReferences(locationsByRoot schemaLocationsByRoot) (resolutions map[*jsonschema.Schema]*jsonschema.Schema, err error) {
	resolutions = map[*jsonschema.Schema]*jsonschema.Schema{}
	for root, locations := range locationsByRoot {
		for schema, location := range locations {
			if schema.Reference != nil {
				ref, err := url.Parse(*schema.Reference)
				if err != nil {
					return nil, errors.WithMessage(err, "failed to parse $ref")
				}

				if location.id != nil {
					// Dereference the $ref against the current base URI
					// (https://tools.ietf.org/html/draft-handrews-json-schema-01#section-8.3.2).
					ref = location.id.URI().ResolveReference(ref)
				}

				// If the $ref's dereferenced URI consists of only a fragment (i.e., it points to a
				// value in the same root schema document), we must not try to resolve it in other
				// root schemas.
				//
				// TODO(sqs): Check that this is correct.
				var onlyInRoot *jsonschema.Schema
				if *ref == (url.URL{Fragment: ref.Fragment}) {
					onlyInRoot = root
				}

				target := resolveReference(ref, locationsByRoot, onlyInRoot)
				if target != nil {
					resolutions[schema] = target
				} else {
					return nil, fmt.Errorf("failed to resolve $ref: %q (dereferenced to %q)", *schema.Reference, ref)
				}
			}
		}
	}
	return resolutions, nil
}

func resolveReference(ref *url.URL, locationsByRoot schemaLocationsByRoot, onlyInRoot *jsonschema.Schema) *jsonschema.Schema {
	if isRefToMetaSchema(ref) {
		return metaSchemaSentinel
	}

	refStr := ref.String()
	for root, locations := range locationsByRoot {
		if onlyInRoot != nil && root != onlyInRoot {
			continue
		}
		for schema, location := range locations {
			if location.id != nil && location.id.String() == refStr {
				return schema
			}
			// TODO(sqs): Eliminate the string hackiness here.
			if onlyInRoot != nil && "/"+jsonschema.EncodeReferenceTokens(location.rel) == ref.Fragment {
				return schema
			}
		}
	}
	return nil
}

// metaSchemaSentinel is a sentinel value that refers to the JSON Schema describing JSON Schema
// documents itself (the meta-schema). During the compiler's resolution phase, it is stored as the
// resolution for $refs to the meta-schema. During the compiler's codegen phase, it is represented
// by the Go type *jsonschema.Schema and an import of package jsonschema is added.
var metaSchemaSentinel = &jsonschema.Schema{}

func isRefToMetaSchema(ref *url.URL) bool {
	return (ref.Scheme == "http" || ref.Scheme == "https") && ref.Host == "json-schema.org" && ref.Path == "/draft-07/schema" && (ref.Fragment == "" || ref.Fragment == "/")
}
