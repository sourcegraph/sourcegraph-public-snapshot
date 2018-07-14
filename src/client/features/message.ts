import {
    MessageActionItem,
    ShowMessageNotification,
    ShowMessageParams,
    ShowMessageRequest,
    ShowMessageRequestParams,
} from '../../protocol'
import { Client } from '../client'
import { StaticFeature } from './common'

/**
 * Support for server messages intended for display to the user (window/showMessages notifications and
 * window/showMessageRequest requests from the server).
 */
export class WindowShowMessageFeature implements StaticFeature {
    constructor(
        private client: Client,
        /** Called when the client receives a window/showMessage notification. */
        private showMessage: (params: ShowMessageParams) => void,
        /**
         * Called when the client receives a window/showMessageRequest request and expected to return a promise
         * that resolves to the selected action.
         */
        private showMessageRequest: (params: ShowMessageRequestParams) => Promise<MessageActionItem | null>
    ) {}

    public initialize(): void {
        // TODO(sqs): no way to unregister these
        this.client.onNotification(ShowMessageNotification.type, this.showMessage)
        this.client.onRequest(ShowMessageRequest.type, this.showMessageRequest)
    }
}
