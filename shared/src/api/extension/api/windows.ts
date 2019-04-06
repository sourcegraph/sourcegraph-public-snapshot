import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { asError } from '../../../util/errors'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { ClientWindowsAPI } from '../../client/api/windows'
import { ViewComponentData } from '../../client/model'
import { ExtCodeEditor } from './codeEditor'
import { ExtDocuments } from './documents'

export interface WindowData {
    visibleViewComponents: ViewComponentData<Pick<sourcegraph.TextDocument, 'uri'>>[]
}

interface WindowsProxyData {
    windows: ClientWindowsAPI
    codeEditor: ClientCodeEditorAPI
}

/**
 * @todo Send the show{Notification,Message,InputBox} requests to the same window (right now they are global).
 * @internal
 */
export class ExtWindow implements sourcegraph.Window {
    private viewComponents = new Map<string, ExtCodeEditor>()

    constructor(private proxy: ProxyResult<WindowsProxyData>, private documents: ExtDocuments, data: WindowData) {
        this._update(data)
    }

    public readonly activeViewComponentChanges = new BehaviorSubject<sourcegraph.ViewComponent | undefined>(undefined)

    public get visibleViewComponents(): sourcegraph.ViewComponent[] {
        const entries = Array.from(this.viewComponents.entries())
        return entries
            .sort(([a], [b]) => {
                if (a < b) {
                    return -1
                }
                if (a > b) {
                    return 1
                }
                return 0
            })
            .map(([, viewComponent]) => viewComponent)
    }

    public get activeViewComponent(): sourcegraph.ViewComponent | undefined {
        return this.activeViewComponentChanges.value
    }

    public showNotification(message: string, type: sourcegraph.NotificationType): void {
        // tslint:disable-next-line: no-floating-promises
        this.proxy.windows.$showNotification(message, type)
    }

    public showMessage(message: string): Promise<void> {
        return this.proxy.windows.$showMessage(message)
    }

    public showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined> {
        return this.proxy.windows.$showInputBox(options)
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
        const reporterProxy = await this.proxy.windows.$showProgress(options)
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

    /**
     * Perform a delta update (update/add/delete) of this window's view components.
     */
    public _update(data: WindowData): void {
        const key = (c: ViewComponentData<Pick<sourcegraph.TextDocument, 'uri'>>): string => `${c.type}:${c.item.uri}`

        const seenKeys = new Set<string>()
        for (const c of data.visibleViewComponents) {
            const k = key(c)
            seenKeys.add(k)
            const existing = this.viewComponents.get(k)
            if (existing) {
                existing._update(c)
            } else {
                this.viewComponents.set(k, new ExtCodeEditor(c, this.proxy.codeEditor, this.documents))
            }
        }
        for (const key of this.viewComponents.keys()) {
            if (!seenKeys.has(key)) {
                this.viewComponents.delete(key)
            }
        }

        // Update active view component.
        const active = data.visibleViewComponents.find(c => c.isActive)
        this.activeViewComponentChanges.next(active ? this.viewComponents.get(key(active)) : undefined)
    }

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI extends ProxyValue {
    $acceptWindowData(win: WindowData | null): void
}

/**
 * @internal
 * @todo Support more than 1 window.
 */
export class ExtWindows implements ExtWindowsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    public activeWindow: ExtWindow | undefined

    /** @internal */
    constructor(
        private proxy: ProxyResult<{ windows: ClientWindowsAPI; codeEditor: ClientCodeEditorAPI }>,
        private documents: ExtDocuments
    ) {}

    public readonly activeWindowChanges = new BehaviorSubject<sourcegraph.Window | undefined>(undefined)

    /**
     * Returns all known windows.
     *
     * @internal
     */
    public getAll(): sourcegraph.Window[] {
        return this.activeWindow ? [this.activeWindow] : []
    }

    /** @internal */
    public $acceptWindowData(win: WindowData | null): void {
        if (win && this.activeWindow) {
            // Update in-place, reusing same object so that object references from extensions to it
            // (and subscriptions to it) remain valid.
            this.activeWindow._update(win)
        } else {
            this.activeWindow = win ? new ExtWindow(this.proxy, this.documents, win) : undefined
            this.activeWindowChanges.next(this.activeWindow)
        }
    }
}
