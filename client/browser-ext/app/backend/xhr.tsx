import { getPlatformName } from "../utils";
import { singleflightFetch } from "./singleflightFetch";

let token: string | null = null;
export function useAccessToken(tok: string): void {
	token = tok;
}

type FetchOptions = { headers: Headers };

export function combineHeaders(a: Headers, b: Headers): Headers {
	let headers = new Headers(a);
	b.forEach((val: string, name: any) => { headers.append(name, val); });
	return headers;
}

function defaultOptions(): FetchOptions | undefined {
	if (typeof Headers === "undefined") {
		return; // for unit tests
	}
	const headers = new Headers();
	if (token) {
		headers.set("Authorization", `session ${token}`);
	}
	headers.set("x-sourcegraph-client", `${getPlatformName()} v${getExtensionVersion()}`);
	return { headers };
};

function getExtensionVersion(): string {
	if (chrome && chrome.runtime && chrome.runtime.getManifest) {
		return chrome.runtime.getManifest().version;
	}
	return "NO_VERSION";
}

const f = singleflightFetch(global.fetch);
export function doFetch(url: string, opt?: any): Promise<Response> {
	let defaults = defaultOptions();
	const fetchOptions = Object.assign({}, defaults, opt);
	if (opt.headers && defaults) {
		// the above object merge might override the auth headers. add those back in.
		fetchOptions.headers = combineHeaders(opt.headers, defaults.headers);
	}
	return f(url, fetchOptions);
}
