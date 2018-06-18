/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Sourcegraph. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
'use strict'

import * as vscode from 'vscode'

export const outputChannel: vscode.OutputChannel = vscode.window.createOutputChannel('LSP')
