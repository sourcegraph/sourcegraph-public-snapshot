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

    /**
     * Gets the active text editor's selection, or the entire file if the selected range is empty.
     */
    getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null

    getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null
    replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void>
    showQuickPick(labels: string[]): Promise<string | undefined>
    showInputBox(options?: InputBoxOptions): Promise<string | undefined>
    showWarningMessage(message: string): Promise<void>
}

export interface InputBoxOptions {
    /**
     * An optional string that represents the title of the input box.
     */
    title?: string

    /**
     * The value to pre-fill in the input box.
     */
    value?: string

    /**
     * Selection of the pre-filled {@linkcode InputBoxOptions.value value}. Defined as tuple of two number where the
     * first is the inclusive start index and the second the exclusive end index. When `undefined` the whole
     * pre-filled value will be selected, when empty (start equals end) only the cursor will be set,
     * otherwise the defined range will be selected.
     */
    valueSelection?: [number, number]

    /**
     * The text to display underneath the input box.
     */
    prompt?: string

    /**
     * An optional string to show as placeholder in the input box to guide the user what to type.
p     */
    placeHolder?: string

    /**
     * Controls if a password input is shown. Password input hides the typed text.
     */
    password?: boolean

    /**
     * Set to `true` to keep the input box open when focus moves to another part of the editor or to another window.
     * This setting is ignored on iPad and is always false.
     */
    ignoreFocusOut?: boolean
}
