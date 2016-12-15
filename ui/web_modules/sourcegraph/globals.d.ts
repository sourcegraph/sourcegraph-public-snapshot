/* tslint:disable */

declare var require: {
	(path: string): any;
	(paths: string[], callback: (...modules: any[]) => void): void;
	ensure: (paths: string[], callback: (...modules: any[]) => void) => void;
};

declare namespace process {
	export var env: { NODE_ENV: string };
}

declare namespace global {
	export var window: any;
	export var document: any;
	export var chrome: any;
	export var __webpack_public_path__: any;
	export var module: any;
	export var __sourcegraphJSContext: any;
}

declare module "expect.js" {
	export default function (arg: any): any;
}

declare module "fuzzysearch" {
	export default {} as any;
}

declare module "react-router-scroll" {
	export function useScroll(arg: any): any;
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

declare module "querystring" {
	export function parse(query: string): any;
}

// declare module "child_process" {
// 	export default {} as any;
// }

// declare namespace Electron {
// 	type CrashReporterStartOptions = any;
// }

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
/**
 * Thenable is a common denominator between ES6 promises, Q, jquery.Deferred, WinJS.Promise,
 * and others. This API makes no assumption about what promise libary is being used which
 * enables reusing existing code without migrating to a specific promise implementation. Still,
 * we recommend the use of native promises which are available in VS Code.
 */
interface Thenable<R> {
	/**
	* Attaches callbacks for the resolution and/or rejection of the Promise.
	* @param onfulfilled The callback to execute when the Promise is resolved.
	* @param onrejected The callback to execute when the Promise is rejected.
	* @returns A Promise for the completion of which ever callback is executed.
	*/
	then<TResult>(onfulfilled?: (value: R) => TResult | Thenable<TResult>, onrejected?: (reason: any) => TResult | Thenable<TResult>): Thenable<TResult>;
	then<TResult>(onfulfilled?: (value: R) => TResult | Thenable<TResult>, onrejected?: (reason: any) => void): Thenable<TResult>;
}
