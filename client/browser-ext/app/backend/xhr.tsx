import "whatwg-fetch";
import { getPlatformName } from "../utils";
import { isBrowserExtension } from "../utils/context";

let token: string | null = null;
export function useAccessToken(tok: string): void {
	token = tok;
}

type FetchOptions = { headers: Headers, credentials: string };

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
	// TODO(uforic): can we get rid of this and just pass cookies instead
	if (isBrowserExtension() && token) {
		headers.set("Authorization", `session ${token}`);
	}
	if (isBrowserExtension()) {
		headers.set("x-sourcegraph-client", `${getPlatformName()} v${getExtensionVersion()}`);
	}
	return {
		headers,
		// we only need to include cookies when running in-page
		// the chrome extension uses the Authorization field
		credentials: isBrowserExtension() ? "omit" : "include",
	};
}

function getExtensionVersion(): string {
	if (chrome && chrome.app && chrome.app.getDetails) {
		const details = chrome.app.getDetails();
		if (details && details.version) {
			return details.version;
		}
	}
	if (chrome && chrome.runtime && chrome.runtime.getManifest) {
		const manifest = chrome.runtime.getManifest();
		if (manifest && manifest.version) {
			return manifest.version;
		}
	}
	return "NO_VERSION";
}

export function doFetch(url: string, opt?: any): Promise<Response> {
	let defaults = defaultOptions();
	const fetchOptions = Object.assign({}, defaults, opt);
	if (opt.headers && defaults) {
		// the above object merge might override the auth headers. add those back in.
		fetchOptions.headers = combineHeaders(opt.headers, defaults.headers);
	}
	return fetch(url, fetchOptions);
}
