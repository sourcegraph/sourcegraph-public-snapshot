// This file imports all of the languages ("modes" in vscode
// terminology) that we support in the UI.

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";
import { deepFreeze } from "sourcegraph/util/deepFreeze";

export const modes = deepFreeze(["c", "go", "ruby", "javascript", "typescript"]);

export const modesToSearch = deepFreeze(modes.filter((mode) => mode !== "javascript"));
