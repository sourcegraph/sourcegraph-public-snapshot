import "vscode";

import { Features } from "sourcegraph/util/features";

// To import and activate a vscode extension, add the following to this file:
//
//   import { activate } from "path/to/extension";
//   activate();

import { activate } from "sourcegraph/ext/lsp/extension";
activate();

import { activate as activateZap } from "sourcegraph/ext/zap/extension";
if (Features.zap.isEnabled()) {
	activateZap();
}