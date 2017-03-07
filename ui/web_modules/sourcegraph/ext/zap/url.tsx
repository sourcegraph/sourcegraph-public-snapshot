// HACK to get/set zap ref in the URL

export function currentZapRef(): string | null {
	const m = document.location.search.match(/(\?|&)tmpZapRef=([^&]+)/);
	if (!m) { return null; }
	return decodeURIComponent(m[2]);
}

export function setCurrentZapRef(ref: string): void {
	const cur = currentZapRef();
	if (cur) {
		document.location.search = document.location.search.replace(`tmpZapRef=${encodeURIComponent(cur)}`, `tmpZapRef=${encodeURIComponent(ref)}`);
	} else {
		let add = document.location.search ? "&" : "?";
		add += `tmpZapRef=${encodeURIComponent(ref)}`;
		document.location.search += add;
	}
}
