import "./fetch";

let token: string | null = null;
export function useAccessToken(tok: string): void {
	token = tok;
}

type FetchOptions = {headers: Headers};

function defaultOptions(): FetchOptions | undefined {
	if (typeof Headers === "undefined") {
		return; // for unit tests
	}

	const headers = new Headers();
	if (token) {
		headers.set("Authorization", `session ${token}`);
	}
	return {headers};
};

export function doFetch(url: string, opt?: Object): Promise<Response> {
	return fetch(url, Object.assign({}, defaultOptions(), opt));
}


