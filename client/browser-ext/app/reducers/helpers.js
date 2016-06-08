export const keyFor = (repo, rev, path, query) => `${repo || null}@${rev || null}@${path || null}@${query}`;

function parseKey(key) {
	const keyRegexp = /(.*)@(.*)@(.*)@(.*)/;
	const result = keyRegexp.exec(key);

	const getValue = (i) => {
	 	if (!result) return null;
	 	if (result[i] === "undefined" || result[i] === "null") return null;
	 	return result[i];
	}

	return {
		repo: getValue(1),
		rev: getValue(2),
		path: getValue(3),
		defPath: getValue(3),
		query: getValue(4),
	};
}

export function getExpiredSrclibDataVersion(data) {
	return Object.keys(data.timestamps)
		.filter((key) => data.timestamps[key] && (Date.now() - data.timestamps[key]) > (1000 * 60)) // expire every min
		.map(parseKey);
}

export function getExpiredDef(data) {
	return Object.keys(data.timestamps)
		.filter((key) => data.timestamps[key] && (Date.now() - data.timestamps[key]) > (1000 * 60 * 24)) // expire every day
		.map(parseKey);
}

export function getExpiredDefs(data) {
	return Object.keys(data.timestamps)
		.filter((key) => data.timestamps[key] && (Date.now() - data.timestamps[key]) > (1000 * 60 * 24)) // expire every day
		.map(parseKey);
}

export function getExpiredAnnotations(data) {
	return Object.keys(data.timestamps)
		.filter((key) => data.timestamps[key] && (Date.now() - data.timestamps[key]) > (1000 * 60 * 24)) // expire every day
		.map(parseKey);
}
