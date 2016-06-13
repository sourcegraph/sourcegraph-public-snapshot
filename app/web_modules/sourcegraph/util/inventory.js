export function getLanguageExtensionForPath(path) {
	let language = null;
	if (path) {
		const parts = path.split("/");
		const lastPart = parts[parts.length - 1];
		const extensionSplit = lastPart.split(".");
		if (extensionSplit.length >= 2) {
			const extension = extensionSplit[extensionSplit.length - 1];
			language = extension.toLowerCase();
		}
	}
	return language;
}
