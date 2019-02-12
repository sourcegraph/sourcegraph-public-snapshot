import { ProxyValue, proxyValueSymbol } from 'comlink'
import { BehaviorSubject, Observer, of } from 'rxjs'
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
    constructor(private windowsProxy: ClientWindowsAPI, private readonly textEditors: ExtCodeEditor[]) {}

    public readonly activeViewComponentChanges = of(this.activeViewComponent)

    public get visibleViewComponents(): sourcegraph.ViewComponent[] {
        return this.textEditors
    }

    public get activeViewComponent(): sourcegraph.ViewComponent | undefined {
        return this.textEditors.find(({ isActive }) => isActive)
    }

    public showNotification(message: string): void {
        this.windowsProxy.$showNotification(message)
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
        const handle = await this.windowsProxy.$startProgress(options)
        const reporter: Observer<sourcegraph.Progress> = {
            next: (progress: sourcegraph.Progress): void => {
                this.windowsProxy.$updateProgress(handle, progress)
            },
            error: (err: any): void => {
                const error = asError(err)
                this.windowsProxy.$updateProgress(handle, undefined, {
                    message: error.message,
                    stack: error.stack,
                })
            },
            complete: (): void => {
                this.windowsProxy.$updateProgress(handle, undefined, undefined, true)
            },
        }
        return reporter
    }

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI {
    $acceptWindowData(allWindows: WindowData[]): void
}

/** @internal */
export class ExtWindows implements ExtWindowsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private data: WindowData[] = []

    /** @internal */
    constructor(
        private windowsProxy: ClientWindowsAPI,
        private codeEditorProxy: ClientCodeEditorAPI,
        private documents: ExtDocuments
    ) {}

    public readonly activeWindowChanged = new BehaviorSubject<sourcegraph.Window | undefined>(this.getActive())

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
                    this.windowsProxy,
                    window.visibleViewComponents.map(
                        c =>
                            new ExtCodeEditor(
                                c.item.uri,
                                c.selections,
                                c.isActive,
                                this.codeEditorProxy,
                                this.documents
                            )
                    )
                )
        )
    }

    /** @internal */
    public $acceptWindowData(allWindows: WindowData[]): void {
        this.data = allWindows
        this.activeWindowChanged.next(this.getActive())
    }
}
