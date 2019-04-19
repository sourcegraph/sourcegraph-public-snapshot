import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { sortBy } from 'lodash'
import { BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { asError } from '../../../util/errors'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { ClientWindowsAPI } from '../../client/api/windows'
import { CodeEditorData, EditorId } from '../../client/services/editorService'
import { ExtCodeEditor } from './codeEditor'
import { ExtDocuments } from './documents'

export interface WindowData {
    editors: readonly (CodeEditorData & EditorId)[]
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
    /** Map of editor key to editor. */
    private viewComponents = new Map<string, ExtCodeEditor>()

    constructor(private proxy: ProxyResult<WindowsProxyData>, private documents: ExtDocuments, data: WindowData) {
        this.update(data)
    }

    public readonly activeViewComponentChanges = new BehaviorSubject<sourcegraph.ViewComponent | undefined>(undefined)

    public get visibleViewComponents(): sourcegraph.ViewComponent[] {
        const entries = Array.from(this.viewComponents.entries())
        return sortBy(entries, 0).map(([, viewComponent]) => viewComponent)
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
    public update(data: WindowData): void {
        const seenEditorIds = new Set<string>()
        for (const c of data.editors) {
            seenEditorIds.add(c.editorId)
            const existing = this.viewComponents.get(c.editorId)
            if (existing) {
                existing.update(c)
            } else {
                this.viewComponents.set(c.editorId, new ExtCodeEditor(c, this.proxy.codeEditor, this.documents))
            }
        }
        for (const editorId of this.viewComponents.keys()) {
            if (!seenEditorIds.has(editorId)) {
                this.viewComponents.delete(editorId)
            }
        }

        // Update active view component.
        const active = data.editors.find(c => c.isActive)
        this.activeViewComponentChanges.next(active ? this.viewComponents.get(active.editorId) : undefined)
    }

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI extends ProxyValue {
    $acceptWindowData(windowData: WindowData | null): void
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
    public $acceptWindowData(windowData: WindowData | null): void {
        if (windowData && this.activeWindow) {
            // Update in-place, reusing same object so that object references from extensions to it
            // (and subscriptions to it) remain valid.
            this.activeWindow.update(windowData)
        } else {
            this.activeWindow = windowData ? new ExtWindow(this.proxy, this.documents, windowData) : undefined
            this.activeWindowChanges.next(this.activeWindow)
        }
    }
}
