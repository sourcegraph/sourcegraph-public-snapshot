import { singleflightFetch } from "./singleflightFetch";

let token: string | null = null;
export function useAccessToken(tok: string): void {
	token = tok;
}

type FetchOptions = { headers: Headers };

function defaultOptions(): FetchOptions | undefined {
	if (typeof Headers === "undefined") {
		return; // for unit tests
	}

	const headers = new Headers();
	if (token) {
		headers.set("Authorization", `session ${token}`);
	}
	headers.set("x-sourcegraph-browser-extension", "true");
	return { headers };
};

const f = singleflightFetch(global.fetch);
export function doFetch(url: string, opt?: any): Promise<Response> {
	return f(url, Object.assign({}, defaultOptions(), opt));
}
