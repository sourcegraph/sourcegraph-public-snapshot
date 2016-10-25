// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";

export const modes = ["c", "go", "ruby", "javascript", "typescript"];
