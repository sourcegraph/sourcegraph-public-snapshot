// tslint:disable

/// <reference path="node_modules/@types/chrome/index.d.ts" />

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
	export var navigator: any;
	export var Promise: any;
    export var Node: any;
}
