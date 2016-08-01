/* tslint:disable: no-namespace no-reference */

/// <reference path="../../node_modules/typescript/lib/lib.es2016.d.ts" />
/// <reference path="../../node_modules/typescript/lib/lib.dom.d.ts" />

declare var require: {
	(path: string): any;
	(paths: string[], callback: (...modules: any[]) => void): void;
	ensure: (paths: string[], callback: (require: (path: string) => any) => void) => void;
};

declare namespace process {
	export var env: {NODE_ENV: string};
}

declare namespace global {
	export var window: any;
	export var it: any; // only set while testing
	export var beforeEach: (f: () => void) => void;
}

declare module "flux/lib/FluxStoreGroup" {
	class FluxStoreGroup {
		constructor(stores: FluxUtils.Store<any>[], callback: () => void);
		release(): void;
	}
	export default FluxStoreGroup;
}

declare module "react-router/lib/PatternUtils" {
	export function matchPattern(pattern: string, pathname: string): {paramNames: string[], paramValues: string[]};
}

declare module "abab" {
	export function atob(b64: string): string;
	export function btoa(str: string): string;
}
