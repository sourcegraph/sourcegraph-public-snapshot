export function authHeaders() {
	let hdr = {};
	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		hdr.authorization = `Basic ${btoa(auth)}`;
	}
	return hdr;
}
