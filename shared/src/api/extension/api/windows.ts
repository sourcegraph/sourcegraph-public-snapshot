import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { BehaviorSubject, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { asError } from '../../../util/errors'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { ClientWindowsAPI } from '../../client/api/windows'
import { ViewComponentData } from '../../client/model'
import { TextDocumentIdentifier } from '../../client/types/textDocument'
import { ExtCodeEditor } from './codeEditor'
import { ExtDocuments } from './documents'

export interface WindowData {
    visibleViewComponents: (Pick<ViewComponentData, 'selections' | 'isActive'> & { item: TextDocumentIdentifier })[]
}

/**
 * @todo Send the show{Notification,Message,InputBox} requests to the same window (right now they are global).
 * @internal
 */
class ExtWindow implements sourcegraph.Window {
    constructor(private windowsProxy: ProxyResult<ClientWindowsAPI>, private readonly textEditors: ExtCodeEditor[]) {}

    public readonly activeViewComponentChanges = of(this.activeViewComponent)

    public get visibleViewComponents(): sourcegraph.ViewComponent[] {
        return this.textEditors
    }

    public get activeViewComponent(): sourcegraph.ViewComponent | undefined {
        return this.textEditors.find(({ isActive }) => isActive)
    }

    public showNotification(message: string, type: sourcegraph.NotificationType): void {
        // tslint:disable-next-line: no-floating-promises
        this.windowsProxy.$showNotification(message, type)
    }

    public showMessage(message: string): Promise<void> {
        return this.windowsProxy.$showMessage(message)
    }

    public showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined> {
        return this.windowsProxy.$showInputBox(options)
    }

    public async withProgress<R>(
        options: sourcegraph.ProgressOptions,
        task: (reporter: sourcegraph.ProgressReporter) => Promise<R>
    ): Promise<R> {
        const reporter = await this.showProgress(options)
        try {
            const result = await task(reporter)
            reporter.complete()
            return result
        } catch (err) {
            reporter.error(err)
            throw err
        }
    }

    public async showProgress(options: sourcegraph.ProgressOptions): Promise<sourcegraph.ProgressReporter> {
        const reporterProxy = await this.windowsProxy.$showProgress(options)
        return {
            next: (progress: sourcegraph.Progress) => {
                // tslint:disable-next-line: no-floating-promises
                reporterProxy.next(progress)
            },
            error: (err: any) => {
                const error = asError(err)
                // tslint:disable-next-line: no-floating-promises
                reporterProxy.error({
                    message: error.message,
                    name: error.name,
                    stack: error.stack,
                })
            },
            complete: () => {
                // tslint:disable-next-line: no-floating-promises
                reporterProxy.complete()
            },
        }
    }

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI extends ProxyValue {
    $acceptWindowData(allWindows: WindowData[]): void
}

/** @internal */
export class ExtWindows implements ExtWindowsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private data: WindowData[] = []

    /** @internal */
    constructor(
        private proxy: ProxyResult<{ windows: ClientWindowsAPI; codeEditor: ClientCodeEditorAPI }>,
        private documents: ExtDocuments
    ) {}

    public readonly activeWindowChanges = new BehaviorSubject<sourcegraph.Window | undefined>(this.getActive())

    /** @internal */
    public getActive(): sourcegraph.Window | undefined {
        return this.getAll()[0]
    }

    /**
     * Returns all known windows.
     *
     * @internal
     */
    public getAll(): sourcegraph.Window[] {
        return this.data.map(
            window =>
                new ExtWindow(
                    this.proxy.windows,
                    window.visibleViewComponents.map(
                        c =>
                            new ExtCodeEditor(
                                c.item.uri,
                                c.selections,
                                c.isActive,
                                this.proxy.codeEditor,
                                this.documents
                            )
                    )
                )
        )
    }

    /** @internal */
    public $acceptWindowData(allWindows: WindowData[]): void {
        this.data = allWindows
        this.activeWindowChanges.next(this.getActive())
    }
}
