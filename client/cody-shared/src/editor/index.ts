export interface ActiveTextEditor {
    content: string
    filePath: string
}

export interface ActiveTextEditorSelection {
    fileName: string
    precedingText: string
    selectedText: string
    followingText: string
}

export interface ActiveTextEditorVisibleContent {
    content: string
    fileName: string
}

export interface Editor {
    getWorkspaceRootPath(): string | null
    getActiveTextEditor(): ActiveTextEditor | null
    getActiveTextEditorSelection(): ActiveTextEditorSelection | null
    getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null
    showQuickPick(labels: string[]): Promise<string | undefined>
    showWarningMessage(message: string): Promise<void>
}
