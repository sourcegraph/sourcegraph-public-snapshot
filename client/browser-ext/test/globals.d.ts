declare var require: {
	(path: string): any;
	(paths: string[], callback: (...modules: any[]) => void): void;
};

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

declare module "nightmare" {
    export class Nightmare<N> {
        constructor(options?: Nightmare.IConstructorOptions);

        // Interact
        goto(url: string): Nightmare<N>;
        back(): Nightmare<N>;
        forward(): Nightmare<N>;
        refresh(): Nightmare<N>;
        click(selector: string): Nightmare<N>;
        type(selector: string, text: string): Nightmare<N>;
        upload(selector: string, path: string): Nightmare<N>;
        scrollTo(top: number, left: number): Nightmare<N>;
        inject(type: string, file: string): Nightmare<N>;
        evaluate<T1, T2, T3, R>(fn: (arg1: T1, arg2: T2, arg3: T3) => R, cb: (result: R) => void, arg1: T1, arg2: T2, arg3: T3): Nightmare<N>;
        evaluate<T1, T2, R>(fn: (arg1: T1, arg2: T2) => R, cb: (result: R) => void, arg1: T1, arg2: T2): Nightmare<N>;
        evaluate<T, R>(fn: (arg: T) => R, cb: (result: R) => void, arg: T): Nightmare<N>;
        evaluate<T>(fn: (arg: T) => void, cb: () => void, arg: T): Nightmare<N>;
        evaluate<R>(fn: () => R, cb: (result: R) => void): Nightmare<N>;
        evaluate<R>(fn: () => R): Nightmare<R>;
        wait(): Nightmare<N>;
        wait(ms: number): Nightmare<N>;
        wait(selector: string): Nightmare<N>;
        wait(fn: () => any, value: any, delay?: number): Nightmare<N>;
        use(plugin: (nightmare: Nightmare<N>) => void): Nightmare<N>;
        run(cb?: (err: any, nightmare: Nightmare<N>) => void): Nightmare<N>;
        end(): Promise<N>;

        // Extract
        exists(selector: string, cb: (result: boolean) => void): Nightmare<N>;
        visible(selector: string, cb: (result: boolean) => void): Nightmare<N>;
        on(event: string, cb: () => void): Nightmare<N>;
        on(event: 'initialized', cb: () => void): Nightmare<N>;
        on(event: 'loadStarted', cb: () => void): Nightmare<N>;
        on(event: 'loadFinished', cb: (status: string) => void): Nightmare<N>;
        on(event: 'urlChanged', cb: (targetUrl: string) => void): Nightmare<N>;
        on(event: 'navigationRequested', cb: (url: string, type: string, willNavigate: boolean, main: boolean) => void): Nightmare<N>;
        on(event: 'resourceRequested', cb: (requestData: Nightmare.IRequest, networkRequest: Nightmare.INetwordRequest) => void): Nightmare<N>;
        on(event: 'resourceReceived', cb: (response: Nightmare.IResponse) => void): Nightmare<N>;
        on(event: 'resourceError', cb: (resourceError: Nightmare.IResourceError) => void): Nightmare<N>;
        on(event: 'consoleMessage', cb: (msg: string, lineNumber: number, sourceId: number) => void): Nightmare<N>;
        on(event: 'alert', cb: (msg: string) => void): Nightmare<N>;
        on(event: 'confirm', cb: (msg: string) => void): Nightmare<N>;
        on(event: 'prompt', cb: (msg: string, defaultValue?: string) => void): Nightmare<N>;
        on(event: 'error', cb: (msg: string, trace?: Nightmare.IStackTrace[]) => void): Nightmare<N>;
        on(event: 'timeout', cb: (msg: string) => void): Nightmare<N>;
        screenshot(path: string): Nightmare<N>;
        pdf(path: string): Nightmare<N>;
        title(cb: (title: string) => void): Nightmare<N>;
        url(cb: (url: string) => void): Nightmare<N>;

        // Settings
        authentication(user: string, password: string): Nightmare<N>;
        useragent(useragent: string): Nightmare<N>;
        viewport(width: number, height: number): Nightmare<N>;
        zoom(zoomFactor: number): Nightmare<N>;
        headers(headers: Object): Nightmare<N>;
    }
}