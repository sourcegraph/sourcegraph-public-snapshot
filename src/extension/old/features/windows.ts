import { BehaviorSubject } from 'rxjs'
import { MessageConnection } from '../../../jsonrpc2/connection'
import {
    DidCloseTextDocumentNotification,
    DidOpenTextDocumentNotification,
    ShowInputRequest,
    TextDocumentDecoration,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
} from '../../../protocol'
import { TextDocumentIdentifier } from '../../../types/textDocument'
import { URI } from '../../../types/uri'
import { Observable, Window, Windows } from '../api'

/**
 * Implements the Sourcegraph extension API's {@link SourcegraphExtensionAPI#windows} value.
 *
 * @param ext The Sourcegraph extension API handle.
 * @return The {@link SourcegraphExtensionAPI#windows} value.
 */
export class ExtWindows extends BehaviorSubject<Window[]> implements Windows, Observable<Window[]> {
    constructor(private rawConnection: MessageConnection) {
        super([
            {
                isActive: true,
                activeComponent: null,
            },
        ])

        // Track last-opened text document.
        rawConnection.onNotification(DidOpenTextDocumentNotification.type, params => {
            this.next([
                { ...this.value[0], activeComponent: { isActive: true, resource: URI.parse(params.textDocument.uri) } },
            ])
        })
        rawConnection.onNotification(DidCloseTextDocumentNotification.type, params => {
            if (
                this.activeWindow &&
                this.activeWindow.activeComponent &&
                this.activeWindow.activeComponent.resource &&
                this.activeWindow.activeComponent.resource.toString() === params.textDocument.uri
            ) {
                this.next([{ ...this.value[0], activeComponent: null }])
            }
        })
    }

    public get activeWindow(): Window | null {
        return this.value.find(({ isActive }) => isActive) || null
    }

    public showInputBox(message: string, defaultValue?: string): Promise<string | null> {
        return this.rawConnection.sendRequest(ShowInputRequest.type, { message, defaultValue })
    }

    public setDecorations(resource: TextDocumentIdentifier, decorations: TextDocumentDecoration[]): void {
        return this.rawConnection.sendNotification(TextDocumentPublishDecorationsNotification.type, {
            textDocument: resource,
            decorations,
        } as TextDocumentPublishDecorationsParams)
    }

    public readonly [Symbol.observable] = () => this
}
