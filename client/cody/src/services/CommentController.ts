import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

import { CodeLensProvider } from './CodeLensProvider'

const initPost = new vscode.Position(0, 0)
const initRange = new vscode.Range(initPost, initPost)

export class CommentController {
    // Controller init
    private readonly id = 'cody-file-chat'
    private readonly label = 'Cody: File Chat'
    private readonly threadLabel = 'Ask Cody...'
    private options = {
        prompt: 'Click here to ask Cody.',
        placeHolder: 'Ask Cody a question, or start with /fix to request edits (e.g. "/fix convert tabs to spaces")',
    }
    public readonly codyIcon: vscode.Uri = this.getIconPath('cody')
    private readonly userIcon: vscode.Uri = this.getIconPath('user')
    private _disposables: vscode.Disposable[] = []

    // Constroller State
    private commentController: vscode.CommentController
    public threads: vscode.CommentReply | null = null // threads contains a thread with comments
    public thread: vscode.CommentThread | null = null // a thread is a comment
    private currentTaskId = ''

    // Editor State
    public editor: vscode.TextEditor | null = null
    public selection: ActiveTextEditorSelection | null = null
    public selectionRange = initRange

    // Doc State
    public isInProgress = false

    // States
    public codeLenses: Map<string, CodeLensProvider> = new Map()

    constructor(private extensionPath: string) {
        this.commentController = vscode.comments.createCommentController(this.id, this.label)
        this.commentController.options = this.options
        // enable controller if feature flag is on
        this.commentController.commentingRangeProvider = {
            provideCommentingRanges: (document: vscode.TextDocument) => {
                const lineCount = document.lineCount
                return [new vscode.Range(0, 0, lineCount - 1, 0)]
            },
        }

        // Track last selection in valid doc
        vscode.window.onDidChangeTextEditorSelection(e => {
            if (e.textEditor.document.uri.scheme !== 'file') {
                return
            }
            const selection = e.selections[0]
            if (selection && !this.isInProgress && this.selectionRange.end.line - 2 !== selection.start.line) {
                const range = new vscode.Range(
                    new vscode.Position(selection.start.line - 1, 0),
                    new vscode.Position(selection.end.line + 2, 0)
                )
                this.selectionRange = range
            }
        })

        // Track and update line of changes when the task for the current selected range is being processed
        vscode.workspace.onDidChangeTextDocument(e => {
            if (!this.isInProgress || !this.selectionRange || e.document.uri.scheme !== 'file') {
                // don't track
                return
            }
            for (const change of e.contentChanges) {
                if (change.range.start.line > this.selectionRange.end.line) {
                    return
                }
                let addedLines = 0
                if (change.text.includes('\n')) {
                    addedLines = change.text.split('\n').length - 1
                } else if (change.range.end.line - change.range.start.line > 0) {
                    addedLines -= change.range.end.line - change.range.start.line
                }
                const newRange = new vscode.Range(
                    new vscode.Position(this.selectionRange.start.line + addedLines, 0),
                    new vscode.Position(this.selectionRange.end.line + addedLines, 0)
                )
                this.selectionRange = newRange
            }
        })
    }

    // Get controller instance
    public get(): vscode.CommentController {
        return this.commentController
    }

    /**
     * Create a new thread (the first comment of a thread)
     */
    public create(humanInput: string): vscode.CommentReply | null {
        const editor = vscode.window.activeTextEditor
        if (!editor || !humanInput || editor.document.uri.scheme !== 'file') {
            return null
        }
        this.thread = this.commentController.createCommentThread(editor?.document.uri, editor.selection, [])
        const threads: vscode.CommentReply = {
            text: humanInput,
            thread: this.thread,
        }
        this.threads = threads
        return threads
    }

    /**
     * List response from Human as comment
     */
    public async chat(threads: vscode.CommentReply, isFixMode: boolean = false): Promise<void> {
        this.isInProgress = true
        const humanInput = threads.text
        const thread = threads.thread
        thread.label = this.threadLabel
        const comment = new Comment(
            this.markdown(humanInput),
            vscode.CommentMode.Preview,
            { name: 'Me', iconPath: this.userIcon },
            isFixMode,
            thread,
            'loading'
        )
        comment.label = `#${comment.id}`
        thread.comments = [...thread.comments, comment]
        // disable reply until the task is completed
        thread.canReply = false
        // store current working docs for diffs after fixup
        if (isFixMode) {
            const currentLens = await this.makeCodeLenses(comment.id, thread.uri, this.extensionPath)
            currentLens.updatePendingStatus(true, thread.range)
            currentLens.decorator.setStatus('pending')
            await currentLens.decorator.decorate(thread.range)
            this.codeLenses.set(comment.id, currentLens)
            this.currentTaskId = comment.id
        }
        this.threads = threads
        this.thread = thread
        this.selection = await this.makeSelection(isFixMode)
        await vscode.commands.executeCommand('setContext', 'cody.replied', false)
    }

    /**
     * List response from Cody as comment
     */
    public reply(text: string): void {
        if (!this.thread) {
            return
        }
        const codyReply = new Comment(
            this.markdown(text),
            vscode.CommentMode.Preview,
            { name: 'Cody', iconPath: this.codyIcon },
            false,
            this.thread,
            undefined
        )
        this.thread.comments = [...this.thread.comments, codyReply]
        this.thread.canReply = true
        this.currentTaskId = ''
        void vscode.commands.executeCommand('setContext', 'cody.replied', true)
    }

    /**
     * Remove a comment thread / conversation
     */
    public delete(thread: vscode.CommentThread): void {
        if (!thread) {
            return
        }
        const comments = thread?.comments as Comment[]
        comments.map(comment => {
            this.codeLenses.get(comment.id)?.remove()
        })
        thread.dispose()
        this.reset()
    }

    public reset(): void {
        this.selectionRange = initRange
        this.thread = null
    }

    public async replaceSelection(replacement: string): Promise<void> {
        const startTime = performance.now()
        const activeEditor = await this.getEditor()
        if (!activeEditor) {
            console.error('Missing editor')
            return
        }
        const chatSelection = this.getSelectionRange()
        const selection = new vscode.Selection(chatSelection.start, new vscode.Position(chatSelection.end.line + 1, 0))
        if (!selection) {
            await vscode.window.showErrorMessage('Missing selection')
            return
        }

        // Stop tracking for file changes to perfotm replacement
        this.isInProgress = false

        // Perform edits
        await activeEditor.edit(edit => {
            edit.replace(selection, replacement)
        })

        const startLine = selection.start.line
        const newLineCount = replacement.split('\n').length - 1
        // Highlight from the start line to the length of the replacement content
        const newRange = new vscode.Range(startLine, 0, startLine + newLineCount, 0)
        await this.setReplacementRange(newRange)

        // check performance time
        console.info('Replacement duration:', performance.now() - startTime)
        this.currentTaskId = ''
        return
    }

    /**
     * Turns string into Markdown string
     */
    private markdown(text: string): vscode.MarkdownString {
        const markdownText = new vscode.MarkdownString(text)
        markdownText.isTrusted = true
        markdownText.supportHtml = true
        return markdownText
    }

    /**
     * When a comment thread is open, the Editor will be switched to the comment input editor.
     * Get the current editor using the comment thread uri instead
     */
    public async getEditor(): Promise<vscode.TextEditor | null> {
        if (!this.thread) {
            return null
        }
        await vscode.window.showTextDocument(this.thread.uri)
        this.editor = vscode.window.activeTextEditor || null
        return this.editor
    }

    public getSelection(): ActiveTextEditorSelection | null {
        return this.selection
    }
    /**
     * Get current selected lines from the comment thread.
     * Add an extra line to the end line to prevent empty selection on single line selection
     */
    public async makeSelection(isFixMode: boolean): Promise<ActiveTextEditorSelection | null> {
        if (!this.thread) {
            return null
        }
        const activeDocument = await vscode.workspace.openTextDocument(this.thread.uri)
        const lineLength = activeDocument.lineAt(this.thread.range.end.line).text.length
        const startPost = new vscode.Position(this.thread.range.start.line, 0)
        const endPostFix = new vscode.Position(this.thread.range.end.line, lineLength)
        const endPostAsk = new vscode.Position(this.thread.range.end.line + 1, 0)
        const selectionRange = new vscode.Range(startPost, isFixMode ? endPostFix : endPostAsk)
        const precedingText = activeDocument.getText(
            new vscode.Range(
                new vscode.Position(Math.max(0, this.thread.range.start.line - SURROUNDING_LINES), 0),
                this.thread.range.start
            )
        )
        const followingText = activeDocument.getText(
            new vscode.Range(
                this.thread.range.end,
                new vscode.Position(this.thread.range.end.line + 1 + SURROUNDING_LINES, 0)
            )
        )
        const selection = {
            fileName: vscode.workspace.asRelativePath(this.thread.uri.fsPath),
            selectedText: activeDocument.getText(selectionRange),
            precedingText,
            followingText,
        }
        this.selectionRange = selectionRange
        this.selection = selection
        return selection
    }

    public getSelectionRange(): vscode.Range {
        return this.selectionRange
    }

    public async makeCodeLenses(taskID: string, uri: vscode.Uri, extPath: string): Promise<CodeLensProvider> {
        const lens = new CodeLensProvider(taskID, extPath)
        const activeDocument = await vscode.workspace.openTextDocument(uri)
        await lens.provideCodeLenses(activeDocument, new vscode.CancellationTokenSource().token)
        vscode.languages.registerCodeLensProvider('*', lens)
        return lens
    }

    private async setReplacementRange(newRange: vscode.Range): Promise<void> {
        this.selectionRange = newRange
        const taskID = this.currentTaskId
        if (this.thread) {
            this.thread.range = newRange
        }
        if (taskID) {
            const lens = this.codeLenses.get(taskID)
            lens?.updatePendingStatus(false, newRange)
            lens?.decorator.setStatus('done')
            await lens?.decorator.decorate(newRange)
        }
    }

    /**
     * Generate icon path for each speaker
     */
    private getIconPath(speaker: string): vscode.Uri {
        const extensionPath = vscode.Uri.file(this.extensionPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')
        return vscode.Uri.joinPath(webviewPath, speaker === 'cody' ? 'cody.png' : 'sourcegraph.png')
    }

    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}

export class Comment implements vscode.Comment {
    public id: string
    public label: string | undefined
    constructor(
        public body: string | vscode.MarkdownString,
        public mode: vscode.CommentMode,
        public author: vscode.CommentAuthorInformation,
        public fixupTask: boolean,
        public parent?: vscode.CommentThread,
        public contextValue?: string,
        public timestamp = new Date(Date.now())
    ) {
        this.id = this.timestamp.getTime().toString()
    }
}
