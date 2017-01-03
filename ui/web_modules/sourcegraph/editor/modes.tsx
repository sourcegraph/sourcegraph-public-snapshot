// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";
import { Features } from "sourcegraph/util/features";

export const modes = new Set<string>(["c", "go", "ruby", "javascript", "typescript"]);

if (Features.langCSS.isEnabled()) {
	modes.add("css");
	modes.add("less");
	modes.add("scss");
}
if (Features.langPHP.isEnabled()) {
	modes.add("php");
}
if (Features.langPython.isEnabled()) {
	modes.add("python");
}
if (Features.langJava.isEnabled()) {
	modes.add("java");
}

export function languagesToSearchModes(languages: string[]): string[] {
	return languages.map(lang => lang.toLowerCase())
		.filter(lang => modes.has(lang) && lang !== "javascript");
}
