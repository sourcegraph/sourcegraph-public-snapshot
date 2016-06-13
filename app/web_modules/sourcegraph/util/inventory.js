export function getLanguageExtensionForPath(path) {
	let language = null;
	if (path) {
		const parts = path.split("/");
		const lastPart = parts[parts.length - 1];
		const extensionSplit = lastPart.split(".");
		if (extensionSplit.length === 2 && lastPart.indexOf(".") !== 0) {
			// don't return extension for dotfiles, like ".gitignore"
			const extension = extensionSplit[1];
			language = extension.toLowerCase();
		}
	}
	return language;
}

export function defPathToLanguage(defPath) {
	if (!defPath) return null;

	if (defPath.startsWith("GoPackage")) return "go";
	if (defPath.startsWith("JavaArtifact")) return "java";

	return null;
}
