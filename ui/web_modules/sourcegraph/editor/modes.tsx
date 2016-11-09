// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";

export const modes = new Set<string>(["c", "go", "ruby", "javascript", "typescript"]);
export const modesToSearch = new Set<string>(["c", "go", "ruby", "typescript"]); // exclude "JavaScript"; backend is the same as TypeScript

export function languagesToSearchModes(languages: string[]): string[] {
	const m: string[] = [];
	languages.forEach((language) => {
		const mode = language.toLowerCase();
		if (modesToSearch.has(mode)) {
			m.push(mode);
		}
	});
	return m;
}
