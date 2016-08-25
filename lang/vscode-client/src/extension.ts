/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
'use strict';

import * as path from 'path';

import { workspace, Disposable, ExtensionContext } from 'vscode';
import { LanguageClient, LanguageClientOptions, SettingMonitor, ServerOptions, ErrorAction, ErrorHandler, CloseAction, TransportKind } from 'vscode-languageclient';

function startLangServer(command: string, documentSelector: string | string[]): Disposable {
	const serverOptions: ServerOptions = {
		command: command,
	};
	const clientOptions: LanguageClientOptions = {
		documentSelector: documentSelector,
	}
	return new LanguageClient(command, serverOptions, clientOptions).start();
}

export function activate(context: ExtensionContext) {
	context.subscriptions.push(startLangServer("sample_server", ["plaintext"]));
	context.subscriptions.push(startLangServer("langserver-go", ["go"]));
}

