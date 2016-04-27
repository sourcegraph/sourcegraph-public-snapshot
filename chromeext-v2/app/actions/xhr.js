let token = null;
export function useAccessToken(tok) {
	token = tok;
}

const defaultOptions = function() {
	const headers = new Headers();
	if (token) {
		let auth = `x-oauth-basic:${token}`;
		headers.set("Authorization", `Basic ${btoa(auth)}`);
	}
	return {headers};
}

const combineHeaders = function (a, b) {
	if (!b) return a;
	if (!a) return b;

	if (!(a instanceof Headers)) throw new Error("must be Headers type");
	if (!(b instanceof Headers)) throw new Error("must be Headers type");

	if (b.forEach) {
		// node-fetch's Headers is not a full implementation and doesn't support iterable,
		// but it does expose forEach.
		b.forEach((val, name) => a.append(name, val));
	} else {
		for (let [name, val] of b) {
			a.append(name, val);
		}
	}
	return a;
}

const checkStatus = function(resp) {
	if (resp.status >= 200 && resp.status <= 299) return resp;
	return resp.text().then((body) => {
		console.error(`HTTP fetch failed with status ${resp.status} ${resp.statusText}: ${resp.url}: ${body}`);
		let err;
		try {
			err = {...(new Error(resp.status)), body: JSON.parse(body)};
		} catch (error) {
			err = {...(new Error(resp.statusText)),
				body: body,
				response: {status: resp.status, statusText: resp.statusText, url: resp.url},
			};
		}
		throw err;
	});
}

export default function(url, init) {
	if (typeof url !== "string") throw new Error("url must be a string (complex requests are not yet supported)");

	const defaults = defaultOptions();

	return fetch(url, {
		...defaults,
		...init,
		headers: combineHeaders(defaults.headers, init ? init.headers : null),
	})
		.then(checkStatus)
		.then((resp) => resp.json());
}


