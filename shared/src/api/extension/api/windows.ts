import * as sourcegraph from 'sourcegraph'
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

    public toJSON(): any {
        return { visibleViewComponents: this.visibleViewComponents, activeViewComponent: this.activeViewComponent }
    }
}

/** @internal */
export interface ExtWindowsAPI {
    $acceptWindowData(allWindows: WindowData[]): void
}

/** @internal */
export class ExtWindows implements ExtWindowsAPI {
    private data: WindowData[] = []

    /** @internal */
    constructor(
        private windowsProxy: ClientWindowsAPI,
        private codeEditorProxy: ClientCodeEditorAPI,
        private documents: ExtDocuments
    ) {}

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
    }
}
