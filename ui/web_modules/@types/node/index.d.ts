// This file exists to satisfy packages that depend on @types/node but
// that we use in such a way that doesn't actually call any Node
// APIs. It is a subset of the index.d.ts file in the npm package
// @types/node.

declare namespace NodeJS { // tslint:disable-line no-namespace
	export type ReadableStream = any;
	export type WritableStream = any;
	export type Process = any;
}

declare module "child_process" { // tslint:disable-line no-namespace
	export type ChildProcess = any;
}
