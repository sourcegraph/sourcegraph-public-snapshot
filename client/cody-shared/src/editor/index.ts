export interface ActiveTextEditor {
    content: string
    filePath: string
    repoName?: string
    revision?: string
}

export interface ActiveTextEditorSelection {
    fileName: string
    repoName?: string
    revision?: string
    precedingText: string
    selectedText: string
    followingText: string
}

export interface ActiveTextEditorVisibleContent {
    content: string
    fileName: string
    repoName?: string
    revision?: string
}

interface VsCodeInlineController {
    selection: ActiveTextEditorSelection | null
}

interface VsCodeTaskContoller {
    add(input: string, selection: ActiveTextEditorSelection): string | null
    stop(taskID: string): void
}

export interface ActiveTextEditorViewControllers {
    inline: VsCodeInlineController
    task: VsCodeTaskContoller
}

export interface Editor {
    controllers?: ActiveTextEditorViewControllers
    getWorkspaceRootPath(): Promise<string | null>
    getActiveTextEditor(): Promise<ActiveTextEditor | null>
    getActiveTextEditorSelection(): Promise<ActiveTextEditorSelection | null>

    /**
     * Gets the active text editor's selection, or the entire file if the selected range is empty.
     */
    getActiveTextEditorSelectionOrEntireFile(): Promise<ActiveTextEditorSelection | null>

    getActiveTextEditorVisibleContent(): Promise<ActiveTextEditorVisibleContent | null>
    replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void>
    showQuickPick(labels: string[]): Promise<string | null>
    showWarningMessage(message: string): Promise<void>
    showInputBox(prompt?: string): Promise<string | null>
}

export class NoopEditor implements Editor {
    public getWorkspaceRootPath(): Promise<string | null> {
        return Promise.resolve(null)
    }

    public getActiveTextEditor(): Promise<ActiveTextEditor | null> {
        return Promise.resolve(null)
    }

    public getActiveTextEditorSelection(): Promise<ActiveTextEditorSelection | null> {
        return Promise.resolve(null)
    }

    public getActiveTextEditorSelectionOrEntireFile(): Promise<ActiveTextEditorSelection | null> {
        return Promise.resolve(null)
    }

    public getActiveTextEditorVisibleContent(): Promise<ActiveTextEditorVisibleContent | null> {
        return Promise.resolve(null)
    }

    public replaceSelection(_fileName: string, _selectedText: string, _replacement: string): Promise<void> {
        return Promise.resolve()
    }

    public showQuickPick(_labels: string[]): Promise<string | null> {
        return Promise.resolve(null)
    }

    public showWarningMessage(_message: string): Promise<void> {
        return Promise.resolve()
    }

    public showInputBox(_prompt?: string): Promise<string | null> {
        return Promise.resolve(null)
    }
}
