// toQuery translates an object into a URL query. It assumes the
// object is shallow (no nested objects).
export default function toQuery(o: Object): string {
	let cmps: Array<string> = [];
	for (let k in o) {
		let v = o[k];
		if (!v) {
			continue;
		}
		if (Array.isArray(v)) {
			for (let i = 0; i < v.length; i++) {
				cmps.push(`${encodeURIComponent(k)}=${encodeURIComponent(v[i])}`);
			}
		} else {
			cmps.push(`${encodeURIComponent(k)}=${encodeURIComponent(v)}`);
		}
	}
	return cmps.join("&");
}
