import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Observable, Subject, Subscription } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ExtWindowsAPI } from '../../extension/api/windows'
import { ViewComponentData } from '../model'
import {
    MessageActionItem,
    MessageType,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../services/notifications'

/** @internal */
export interface ClientWindowsAPI extends ProxyValue {
    $showNotification(message: string): void
    $showMessage(message: string): Promise<void>
    $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined>
    $showProgress(options: sourcegraph.ProgressOptions): sourcegraph.ProgressReporter & ProxyValue
}

/** @internal */
export class ClientWindows implements ClientWindowsAPI {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()

    constructor(
        private proxy: ProxyResult<ExtWindowsAPI>,
        modelVisibleViewComponents: Observable<ViewComponentData[] | null>,
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
        private showInput: (params: ShowInputParams) => Promise<string | null>,
        private createProgressReporter: (options: sourcegraph.ProgressOptions) => Subject<sourcegraph.Progress>
    ) {
        this.subscriptions.add(
            modelVisibleViewComponents.subscribe(viewComponents => {
                // tslint:disable-next-line: no-floating-promises
                this.proxy.$acceptWindowData(
                    viewComponents
                        ? [
                              {
                                  visibleViewComponents: viewComponents.map(viewComponent => ({
                                      item: viewComponent.item,
                                      selections: viewComponent.selections,
                                      isActive: viewComponent.isActive,
                                  })),
                              },
                          ]
                        : []
                )
            })
        )
    }

    public $showNotification(message: string): void {
        this.showMessage({ type: MessageType.Info, message })
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
        }).then(v =>
            // TODO(sqs): update the showInput API to unify null/undefined etc between the old internal API and the new
            // external API.
            v === null ? undefined : v
        )
    }

    public $showProgress(options: sourcegraph.ProgressOptions): sourcegraph.ProgressReporter & ProxyValue {
        return proxyValue(this.createProgressReporter(options))
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
