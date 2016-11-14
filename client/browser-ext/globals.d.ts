/// <reference path="node_modules/@types/chrome/index.d.ts" />

declare var require: {
	(path: string): any;
	(paths: string[], callback: (...modules: any[]) => void): void;
}

declare namespace process {
	export var env: {NODE_ENV: string};
}

declare namespace global {
	export var window: any;
	export var document: any;
	export var navigator: any;
	export var Promise: any;
}

declare namespace Nightmare {
    export interface IConstructorOptions {
        timeout?: any;  // number | string;
        interval?: any; // number | string;
        port?: number;
        weak?: boolean;
        loadImages?: boolean;
        ignoreSslErrors?: boolean;
        sslProtocol?: string;
        webSecurity?: boolean;
        proxy?: string;
        proxyType?: string;
        proxyAuth?: string;
        cookiesFile?: string;
        phantomPath?: string;
        show?: boolean;
    }

    export interface IRequest {
        id: number;
        method: string;
        url: string;
        time: Date;
        headers: Object;
    }
    export interface INetwordRequest {
        abort(): void;
        changeUrl(url: string): void;
        setHeader(key: string, value: string): void;
    }
    export interface IResponse {
        id: number;
        url: string;
        time: Date;
        headers: Object;
        bodySize: number;
        contentType: string;
        redirectURL: string;
        stage: string;
        status: number;
        statusText: string;
    }
    export interface IResourceError {
        id: number;
        url: string;
        errorCode: number;
        errorString: string;
    }
    export interface IStackTrace {
        file: string;
        line: number;
        function?: string;
    }
}