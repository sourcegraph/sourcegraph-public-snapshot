// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";

export const modes = new Set<string>(["c", "go", "ruby", "javascript", "typescript", "python"]);

export function languagesToSearchModes(languages: string[]): string[] {
	const m = new Set<string>();
	languages.forEach((language) => {
		const mode = language.toLowerCase();
		if (modes.has(mode)) {
			m.add(mode);
		}
	});

	if (m.has("javascript") && m.has("typescript")) {
		m.delete("javascript");
	}

	return Array.from(m);
}
