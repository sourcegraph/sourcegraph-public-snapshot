import "./fetch";

let token = null;
export function useAccessToken(tok) {
	token = tok;
}

const defaultOptions = function() {
	if (typeof Headers === "undefined") {
		return; // for unit tests
	}

	const headers = new Headers();
	if (token) {
		headers.set("Authorization", `session ${token}`);
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
		throw {...(new Error(resp.statusText)),
			response: {status: resp.status, url: resp.url},
		};
	});
}

export default function(url, init) {
	if (typeof url !== "string") throw new Error("url must be a string (complex requests are not yet supported)");

	const defaults = defaultOptions();

	return fetch(url, {
		...defaults,
		...init,
		headers: combineHeaders(defaults ? defaults.headers : null, init && init.headers ? init.headers : null),
	})
		.then(checkStatus)
		.then((resp) => resp.json());
}


