import { BehaviorSubject } from 'rxjs'
import { TextDocumentIdentifier } from 'vscode-languageserver-types'
import {
    DidCloseTextDocumentNotification,
    DidOpenTextDocumentNotification,
    ShowInputRequest,
    TextDocumentDecoration,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
} from '../../protocol'
import { Observable, SourcegraphExtensionAPI, Window, Windows } from '../api'

/**
 * Implements the Sourcegraph extension API's {@link SourcegraphExtensionAPI#windows} value.
 *
 * @param ext The Sourcegraph extension API handle.
 * @return The {@link SourcegraphExtensionAPI#windows} value.
 */
export class ExtWindows extends BehaviorSubject<Window[]> implements Windows, Observable<Window[]> {
    constructor(private ext: Pick<SourcegraphExtensionAPI<any>, 'rawConnection'>) {
        super([
            {
                isActive: true,
                activeComponent: null,
            },
        ])

        // Track last-opened text document.
        ext.rawConnection.onNotification(DidOpenTextDocumentNotification.type, params => {
            this.next([{ ...this.value[0], activeComponent: { isActive: true, resource: params.textDocument.uri } }])
        })
        ext.rawConnection.onNotification(DidCloseTextDocumentNotification.type, params => {
            if (
                this.activeWindow &&
                this.activeWindow.activeComponent &&
                this.activeWindow.activeComponent.resource &&
                this.activeWindow.activeComponent.resource === params.textDocument.uri
            ) {
                this.next([{ ...this.value[0], activeComponent: null }])
            }
        })
    }

    public get activeWindow(): Window | null {
        return this.value.find(({ isActive }) => isActive) || null
    }

    public showInputBox(message: string, defaultValue?: string): Promise<string | null> {
        return this.ext.rawConnection.sendRequest(ShowInputRequest.type, { message, defaultValue })
    }

    public setDecorations(resource: TextDocumentIdentifier, decorations: TextDocumentDecoration[]): void {
        return this.ext.rawConnection.sendNotification(TextDocumentPublishDecorationsNotification.type, {
            textDocument: resource,
            decorations,
        } as TextDocumentPublishDecorationsParams)
    }

    public readonly [Symbol.observable] = () => this
}
