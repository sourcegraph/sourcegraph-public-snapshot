// Package feature contains a simple implementation of feature toggles.
//
// For example if there is a feature called Discussions, you can enable it by
// prefixing your src serve command with SG_FEATURE_DISCUSSIONS=t.
//
// To add a new feature flag just update the struct Features in feature.go. To
// check in the Go code just use a simple if statement:
//
//  import "sourcegraph.com/sourcegraph/pkg/conf/feature"
//
//  if feature.Features.MyFeature {
//    ...
//  }
//
// For templates we expose feature.Features struct as Features. Similarly in JS we
// have a variable called Features in the globals module.
//
// Rationale
//
// Our feature implementation is intentionally simple, limited in scope and
// static. We limit the scope to gating experimental features which can be
// opted into at the start of the src process. We make the features static so
// we they are easy to use and reason about. Additionally the compiler will
// catch issues around uses of the feature flag.
package feature
