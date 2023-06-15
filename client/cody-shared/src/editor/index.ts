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
    error(): Promise<void>
}

interface VsCodeFixupController {
    getTaskRecipeData(taskId: string): Promise<
        | {
              instruction: string
              fileName: string
              precedingText: string
              selectedText: string
              followingText: string
          }
        | undefined
    >
}

export interface ActiveTextEditorViewControllers {
    inline: VsCodeInlineController
    fixups: VsCodeFixupController
}

export interface Editor {
    controllers?: ActiveTextEditorViewControllers
    getWorkspaceRootPath(): string | null
    getActiveTextEditor(): ActiveTextEditor | null
    getActiveTextEditorSelection(): ActiveTextEditorSelection | null

    /**
     * Gets the active text editor's selection, or the entire file if the selected range is empty.
     */
    getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null

    getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null
    replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void>
    showQuickPick(labels: string[]): Promise<string | undefined>
    showWarningMessage(message: string): Promise<void>
    showInputBox(prompt?: string): Promise<string | undefined>

    // TODO: When Non-Stop Fixup doesn't depend directly on the chat view,
    // move the recipe to client/cody and remove this entrypoint.
    didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void>
}

export class NoopEditor implements Editor {
    public getWorkspaceRootPath(): string | null {
        return null
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        return null
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        return null
    }

    public getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null {
        return null
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        return null
    }

    public replaceSelection(_fileName: string, _selectedText: string, _replacement: string): Promise<void> {
        return Promise.resolve()
    }

    public showQuickPick(_labels: string[]): Promise<string | undefined> {
        return Promise.resolve(undefined)
    }

    public showWarningMessage(_message: string): Promise<void> {
        return Promise.resolve()
    }

    public showInputBox(_prompt?: string): Promise<string | undefined> {
        return Promise.resolve(undefined)
    }

    public didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        return Promise.resolve()
    }
}
