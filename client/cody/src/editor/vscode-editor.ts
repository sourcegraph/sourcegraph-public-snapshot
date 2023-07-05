import * as vscode from 'vscode'

import {
    Editor,
    Indentation,
    LightTextDocument,
    TextDocument,
    TextEdit,
    Uri,
    Workspace,
} from '@sourcegraph/cody-shared/src/editor'

import { FixupController } from '../non-stop/FixupController'
import { InlineController } from '../services/InlineController'

export class VSCodeEditor extends Editor {
    constructor(
        public controllers: {
            inline: InlineController
            fixups: FixupController
        }
    ) {
        super()
    }

    // TODO: Bad internet so ironically enough I can't use Sourcegraph to figure this one out atm
    // public get fileName(): string {
    //     return vscode.window.activeTextEditor?.document.fileName ?? ''
    // }

    public getActiveWorkspace(): Workspace | null {
        const getRoot = (): string | null => {
            const uri = vscode.window.activeTextEditor?.document?.uri
            if (uri) {
                const wsFolder = vscode.workspace.getWorkspaceFolder(uri)
                if (wsFolder) {
                    return wsFolder.uri.toString()
                }
            }
            return vscode.workspace.workspaceFolders?.[0]?.uri?.toString() ?? null
        }

        const root = getRoot()
        return root ? new Workspace(root) : null
    }

    public getWorkspaceOf(uri: Uri): Workspace | null {
        const wsFolder = vscode.workspace.getWorkspaceFolder(vscode.Uri.parse(uri))

        if (wsFolder) {
            return new Workspace(wsFolder.uri.toString())
        }

        return null
    }

    public getActiveLightTextDocument(): LightTextDocument | null {
        const activeEditor = this.getActiveTextEditorInstance()
        if (!activeEditor) {
            return null
        }

        return {
            uri: activeEditor.document.uri.toString(),
            languageId: activeEditor.document.languageId,
        }
    }

    public getOpenLightTextDocuments(): LightTextDocument[] {
        return vscode.workspace.textDocuments.map(doc => ({
            uri: doc.uri.toString(),
            languageId: doc.languageId,
        }))
    }

    private getActiveTextEditorInstance(): vscode.TextEditor | null {
        const activeEditor = vscode.window.activeTextEditor
        return activeEditor && activeEditor.document.uri.scheme === 'file' ? activeEditor : null
    }

    public async getLightTextDocument(uri: Uri): Promise<LightTextDocument | null> {
        const document = await vscode.workspace.openTextDocument(uri)

        if (!document) {
            return null
        }

        return {
            uri,
            languageId: document.languageId,
        }
    }

    public async getTextDocument(uri: Uri): Promise<TextDocument | null> {
        const activeEditor = this.getActiveTextEditorInstance()
        const document = await vscode.workspace.openTextDocument(uri)
        const isActiveEditor = activeEditor && document.uri.toString() === activeEditor.document.uri.toString()

        if (!document) {
            return null
        }

        const visibleRange = isActiveEditor ? activeEditor.visibleRanges[0] : null

        const selection = isActiveEditor ? activeEditor.selection : null

        return {
            uri,
            languageId: document.languageId,

            content: document.getText(),

            visible: visibleRange
                ? {
                      position: visibleRange,
                      offset: null,
                  }
                : null,
            selection:
                selection && !selection.isEmpty
                    ? {
                          position: selection,
                          offset: null,
                      }
                    : null,

            // TODO
            repoName: null,
            revision: null,
        }
    }

    public async replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void> {
        const activeEditor = this.getActiveTextEditorInstance()
        if (this.controllers.inline.isInProgress) {
            await this.controllers.inline.replace(fileName, replacement, selectedText)
            return
        }
        if (!activeEditor || vscode.workspace.asRelativePath(activeEditor.document.uri.fsPath) !== fileName) {
            // TODO: should return something indicating success or failure
            console.error('Missing file')
            return
        }
        const selection = activeEditor.selection
        if (!selection) {
            console.error('Missing selection')
            return
        }
        if (activeEditor.document.getText(selection) !== selectedText) {
            // TODO: Be robust to this.
            await vscode.window.showInformationMessage(
                'The selection changed while Cody was working. The text will not be edited.'
            )
            return
        }

        // Editing the document
        await activeEditor.edit(edit => {
            edit.replace(selection, replacement)
        })

        return
    }

    public async quickPick(labels: string[]): Promise<string | null> {
        const label = await vscode.window.showQuickPick(labels)
        return label ?? null
    }

    public async warn(message: string): Promise<void> {
        await vscode.window.showWarningMessage(message)
    }

    public async prompt(prompt?: string): Promise<string | null> {
        return (
            (await vscode.window.showInputBox({
                placeHolder: prompt || 'Enter here...',
            })) ?? null
        )
    }

    public getIndentation(): Indentation {
        return vscode.window.activeTextEditor
            ? {
                  kind: vscode.window.activeTextEditor.options.insertSpaces ? 'space' : 'tab',
                  size:
                      // tabSize is always resolved to a number when accessing the property
                      vscode.window.activeTextEditor.options.tabSize as number,
              }
            : {
                  kind: 'space',
                  size: 2,
              }
    }

    public edit(uri: string, edits: TextEdit[]): Promise<void> {
        throw new Error('TODO: implement edit for vscode')
    }

    // TODO: When Non-Stop Fixup doesn't depend directly on the chat view,
    // move the recipe to client/cody and remove this entrypoint.
    public async didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        await this.controllers.fixups.didReceiveFixupText(id, text, state)
    }
}
