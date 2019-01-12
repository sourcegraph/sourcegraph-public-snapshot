import { Observable, Subject, Subscription } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtWindowsAPI } from '../../extension/api/windows'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { ViewComponentData } from '../model'
import {
    MessageActionItem,
    MessageType,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../services/notifications'
import { SubscriptionMap } from './common'

/** @internal */
export interface ClientWindowsAPI {
    $showNotification(message: string): void
    $showMessage(message: string): Promise<void>
    $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined>
    $startProgress(options: sourcegraph.ProgressOptions): Promise<number>
    $updateProgress(handle: number, progress?: sourcegraph.Progress, error?: any, done?: boolean): void
}

/** @internal */
export class ClientWindows implements ClientWindowsAPI {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtWindowsAPI

    constructor(
        connection: Connection,
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
        this.proxy = createProxyAndHandleRequests('windows', connection, this)

        this.subscriptions.add(
            modelVisibleViewComponents.subscribe(viewComponents => {
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
        }).then(v =>
            // TODO(sqs): update the showInput API to unify null/undefined etc between the old internal API and the new
            // external API.
            v === null ? undefined : v
        )
    }

    private handles = 1
    private progressReporters = new Map<number, Subject<sourcegraph.Progress>>()

    public async $startProgress(options: sourcegraph.ProgressOptions): Promise<number> {
        const handle = this.handles++
        const reporter = this.createProgressReporter(options)
        this.progressReporters.set(handle, reporter)
        return handle
    }

    public $updateProgress(handle: number, progress?: sourcegraph.Progress, error?: any, done?: boolean): void {
        const reporter = this.progressReporters.get(handle)
        if (!reporter) {
            console.warn('No ProgressReporter for handle ' + handle)
            return
        }
        if (done || (progress && progress.percentage && progress.percentage >= 100)) {
            reporter.complete()
        } else if (error) {
            reporter.error(error)
        } else if (progress) {
            reporter.next(progress)
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
