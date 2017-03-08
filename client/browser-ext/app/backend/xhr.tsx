import { getPlatformName } from "../utils";
import { phabricatorInstance } from "../utils/context";

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
	if (token) {
		headers.set("Authorization", `session ${token}`);
	}
	if (!phabricatorInstance) {
		headers.set("x-sourcegraph-client", `${getPlatformName()} v${getExtensionVersion()}`);
	}
	return {
		headers,
		credentials: "include",
	};
};

function getExtensionVersion(): string {
	if (chrome && chrome.runtime && chrome.runtime.getManifest) {
		return chrome.runtime.getManifest().version;
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
	return global.fetch(url, fetchOptions);
}
