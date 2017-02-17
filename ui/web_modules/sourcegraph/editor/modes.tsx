// TODO(john): this file is defunct and the monaco imports should be moved.
// The languagesToSearchModes will be removed (and the TODO completed) when we
// replace our quickopen w/ vscode's.

// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";
import { getModes } from "sourcegraph/util/features";

const modes = getModes();

export function languagesToSearchModes(languages: string[]): string[] {
	return languages.map(lang => lang.toLowerCase())
		.filter(lang => modes.has(lang) && lang !== "javascript");
}
