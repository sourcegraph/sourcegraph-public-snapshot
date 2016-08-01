declare function proxy(options: proxy.Options): (urlStr: string) => void;
declare namespace proxy {
    interface Options {
        proxy?: string;
        httpProxy?: string;
        httpsProxy?: string;
        noProxy?: string;
    }
}
export = proxy;
