# Upgrade Guide

## v0.7.x to 0.8.x

This release has two breaking changes:

1. If you are using TypeScript 1.7+ and specify `target: es6` and
`module: commonjs`, the output will now be CommonJS instead of ES6
modules. This brings ts-loader into alignment with `tsc`.
2. Declaration files are now emitted when `declaration: true` is 
specified in the tsconfig.json.

## v0.6.x to 0.7.x

This release changed loader messages to print on stderr instead of
stdout. While this shouldn't affect most, if for some reason you relied
on messages going to stdout or on messages *not* going to stderr you
may need to make a change.

## v0.5.x to v0.6.x

This release removed support for TypeScript 1.5 and adds preliminary
support for TypeScript 1.7. Please upgrade to the stable release of 
TypeScript 1.6 or above.

## v0.4.x to v0.5.x

This release removed the dependency on TypeScript from the loader. This
was done so that it's very easy to use the nightly version of TypeScript
by installing `typescript@next`. This does mean that you are responsible
for installing TypeScript yourself.

## v0.3.x to v0.4.x

This release added support for TypeScript 1.5. One of the major changes
introduced in TypeScript 1.5 is the 
[tsconfig.json](https://github.com/Microsoft/TypeScript/wiki/tsconfig.json)
file. All of the TypeScript options that were previously defined through
the loader querystring (`module`, `target`, etc) should now be specified
in the tsconfig.json file instead. The querystring options have been
removed. 