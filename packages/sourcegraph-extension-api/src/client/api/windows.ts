import { Observable, Subscription } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtWindowsAPI } from '../../extension/api/windows'
import {
    MessageActionItem,
    MessageType,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { TextDocumentItem } from '../types/textDocument'
import { SubscriptionMap } from './common'

/** @internal */
export interface ClientWindowsAPI {
    $showNotification(message: string): void
    $showMessage(message: string): Promise<void>
    $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined>
}

/** @internal */
export class ClientWindows implements ClientWindowsAPI {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtWindowsAPI

    constructor(
        connection: Connection,
        environmentTextDocuments: Observable<TextDocumentItem[] | null>,
        /** Called when the client receives a window/showMessage notification. */
        private showMessage: (params: ShowMessageParams) => void,
        /**
         * Called when the client receives a window/showMessageRequest request and expected to return a promise
         * that resolves to the selected action.
         */
        private showMessageRequest: (params: ShowMessageRequestParams) => Promise<MessageActionItem | null>,
        /**
         * Called when the client receives a window/showInput request and expected to return a promise that
         * resolves to the user's input.
         */
        private showInput: (params: ShowInputParams) => Promise<string | null>
    ) {
        this.proxy = createProxyAndHandleRequests('windows', connection, this)

        this.subscriptions.add(
            environmentTextDocuments.subscribe(textDocuments => {
                this.proxy.$acceptWindowData(
                    textDocuments ? textDocuments.map(textDocument => ({ visibleTextDocument: textDocument.uri })) : []
                )
            })
        )

        this.subscriptions.add(this.registrations)
    }

    public $showNotification(message: string): void {
        return this.showMessage({ type: MessageType.Info, message })
    }

    public $showMessage(message: string): Promise<void> {
        return this.showMessageRequest({ type: MessageType.Info, message }).then(
            v =>
                // TODO(sqs): update the showInput API to unify null/undefined etc between the old internal API and the new
                // external API.
                undefined
        )
    }

    public $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined> {
        return this.showInput({
            message: options && options.prompt ? options.prompt : '',
            defaultValue: options && options.value,
        }).then(
            v =>
                // TODO(sqs): update the showInput API to unify null/undefined etc between the old internal API and the new
                // external API.
                v === null ? undefined : v
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
