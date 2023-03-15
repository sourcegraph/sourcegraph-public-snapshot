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

export interface Editor {
    getActiveTextEditor(): ActiveTextEditor | null
    getActiveTextEditorSelection(): ActiveTextEditorSelection | null
    showQuickPick(labels: string[]): Promise<string | undefined>
    showWarningMessage(message: string): Promise<void>
}
