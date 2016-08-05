/* tslint:disable */

/// <reference path="../../node_modules/typescript/lib/lib.es2016.d.ts" />
/// <reference path="../../node_modules/typescript/lib/lib.dom.d.ts" />

declare var require: {
	(path: string): any;
	(paths: string[], callback: (...modules: any[]) => void): void;
};

declare namespace process {
	export var env: {NODE_ENV: string};
}

declare namespace global {
	export var window: any;
	export var document: any;
	export var chrome: any;
	export var it: any; // only set while testing
	export var beforeEach: (f: () => void) => void;
	export var setTimeout: any;
	export var __webpack_public_path__: any;
	export var module: any;
}

declare module "flux/lib/FluxStoreGroup" {
	class FluxStoreGroup {
		constructor(stores: FluxUtils.Store<any>[], callback: () => void);
		release(): void;
	}
	export default FluxStoreGroup;
}

declare module "react-router" {
	export var applyRouterMiddleware: any;
}

declare module "react-router/lib/PatternUtils" {
	export function matchPattern(pattern: string, pathname: string): {paramNames: string[], paramValues: string[]};
}

declare module "react-router/lib/matchRoutes" {
	export default function matchRoutes(routes: any, location: any, callback: any, remainingPathname?: any); 
}

declare module "abab" {
	export function atob(b64: string): string;
	export function btoa(str: string): string;
}

declare module "react-css-modules" {
	export default function<P>(cls: P, styles: any, options?: any): P;
}

declare module "lodash.debounce" {
	export default _.debounce;
}

declare module "lodash.get" {
	export default _.get;
}

declare module "lodash.uniq" {
	export default _.uniq;
}

declare module "lodash.trimleft" {
	export default _.trimStart;
}

declare module "recharts" {
	export var LineChart: any;
	export var Line: any;
	export var XAxis: any;
	export var YAxis: any;
	export var CartesianGrid: any;
	export var Tooltip: any;
	export var Legend: any;
	export var ReferenceLine: any;
}

declare namespace JSX {
	interface IntrinsicAttributes {
		styleName?: string,
	}
}

declare module "expect.js" {
	export default function(arg: any): any;
}

declare module "fuzzysearch" {
	export default {} as any;
}

declare module "react/lib/update" {
	export default {} as any;
}

declare module "react-router-scroll" {
	export default {} as any;
}

declare module "react-helmet" {
	export default {} as any;
}

declare module "react-hot-loader" {
	export var AppContainer: any;
}

declare module "redbox-react" {
	export default {} as any;
}

declare module "fs" {
	export default {} as any;
}

declare module "child_process" {
	export default {} as any;
}

declare module "react/lib/CSSPropertyOperations" {
	export default {} as any;
}

declare module "react/lib/shallowCompare" {
	export default {} as any;
}
