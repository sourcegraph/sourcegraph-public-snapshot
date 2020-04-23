import { Remote, ProxyMarked, proxyMarker } from '@sourcegraph/comlink'
import { sortBy } from 'lodash'
import { BehaviorSubject, Observable, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { asError } from '../../../util/errors'
import { ClientCodeEditorAPI } from '../../client/api/codeEditor'
import { ClientWindowsAPI } from '../../client/api/windows'
import { EditorUpdate } from '../../client/services/editorService'
import { ExtCodeEditor } from './codeEditor'
import { ExtDocuments } from './documents'

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

    constructor(private proxy: Remote<WindowsProxyData>, private documents: ExtDocuments, data: EditorUpdate[]) {
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
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
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
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                reporterProxy.next(progress)
            },
            error: (err: any) => {
                const error = asError(err)
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                reporterProxy.error({
                    message: error.message,
                    name: error.name,
                    stack: error.stack,
                })
            },
            complete: () => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                reporterProxy.complete()
            },
        }
    }

    /**
     * Perform a delta update (update/add/delete) of this window's view components.
     */
    public update(data: EditorUpdate[]): void {
        for (const update of data) {
            const { editorId } = update
            switch (update.type) {
                case 'added': {
                    const editor = new ExtCodeEditor(
                        { editorId, ...update.editorData },
                        this.proxy.codeEditor,
                        this.documents
                    )
                    this.viewComponents.set(editorId, editor)
                    if (update.editorData.isActive) {
                        this.activeViewComponentChanges.next(editor)
                    }
                    break
                }
                case 'updated': {
                    const editor = this.viewComponents.get(editorId)
                    if (!editor) {
                        throw new Error(`Could not perform update: editor ${editorId} not found`)
                    }
                    editor.update(update.editorData)
                    break
                }
                case 'deleted': {
                    this.viewComponents.delete(editorId)
                    break
                }
            }
        }
    }

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI extends ProxyMarked {
    $acceptWindowData(editorUpdates: EditorUpdate[]): void
}

/**
 * @internal
 * @todo Support more than 1 window.
 */
export class ExtWindows implements ExtWindowsAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    public activeWindow: ExtWindow
    public readonly activeWindowChanges: Observable<sourcegraph.Window>

    /** @internal */
    constructor(
        private proxy: Remote<{ windows: ClientWindowsAPI; codeEditor: ClientCodeEditorAPI }>,
        private documents: ExtDocuments
    ) {
        this.activeWindow = new ExtWindow(this.proxy, this.documents, [])
        this.activeWindowChanges = of(this.activeWindow)
    }

    /**
     * Returns all known windows.
     *
     * @internal
     */
    public getAll(): sourcegraph.Window[] {
        return [this.activeWindow]
    }

    /** @internal */
    public $acceptWindowData(editorUpdates: EditorUpdate[]): void {
        this.activeWindow.update(editorUpdates)
    }
}
