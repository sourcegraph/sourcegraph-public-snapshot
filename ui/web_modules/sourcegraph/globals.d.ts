/* tslint:disable */

/// <reference path="../../node_modules/monaco-editor/monaco.d.ts" />

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
	export var __webpack_public_path__: any;
	export var module: any;
}

declare module "expect.js" {
	export default function(arg: any): any;
}

declare module "fuzzysearch" {
	export default {} as any;
}

declare module "react-router-scroll" {
	export function useScroll(arg: any): any;
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

declare module "url" {
	export function parse(url: string, parseQueryString?: boolean): any;
}

declare module "child_process" {
	export default {} as any;
}
