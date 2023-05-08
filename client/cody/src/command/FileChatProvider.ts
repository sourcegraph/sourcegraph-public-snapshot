import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

import { CodeLensProvider } from './CodeLensProvider'
import { CodyContentProvider } from './ContentProvider'

export class FileChatMessage implements vscode.Comment {
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

const initPost = new vscode.Position(0, 0)
const initRange = new vscode.Range(initPost, initPost)

export class FileChatProvider {
    // Controller init
    private readonly id = 'cody-file-chat'
    private readonly label = 'Cody: File Chat'
    private readonly threadLabel = 'Ask Cody...'
    private options = {
        prompt: 'Reply...',
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
    public addedLines = 0
    public isInProgress = false
    public diffFiles: string[] = []

    // States
    public codeLenses: Map<string, CodeLensProvider> = new Map()

    constructor(
        private extensionPath: string,
        public contentProvider: CodyContentProvider,
        private enableController: boolean
    ) {
        this.commentController = vscode.comments.createCommentController(this.id, this.label)
        this.commentController.options = this.options
        // enable controller if feature flag is on
        if (this.enableController) {
            this.commentController.commentingRangeProvider = {
                provideCommentingRanges: (document: vscode.TextDocument) => {
                    const lineCount = document.lineCount
                    return [new vscode.Range(0, 0, lineCount - 1, 0)]
                },
            }
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
                this.addedLines = addedLines
            }
        })
    }

    // Get controller instance
    public get(): vscode.CommentController {
        return this.commentController
    }

    /**
     * Create a new thread
     */
    public newThreads(humanInput: string): vscode.CommentReply | null {
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
        this.addedLines = 0
        const humanInput = threads.text
        const thread = threads.thread
        thread.label = this.threadLabel
        const newComment = new FileChatMessage(
            this.markdown(humanInput),
            vscode.CommentMode.Preview,
            { name: 'Me', iconPath: this.userIcon },
            isFixMode,
            thread,
            'loading'
        )
        thread.comments = [...thread.comments, newComment]
        // disable reply until the task is completed
        thread.canReply = false
        // store current working docs for diffs after fixup
        if (isFixMode) {
            newComment.label = `#${newComment.id}`
            await this.contentProvider.set(thread.uri, newComment.id)
            this.diffFiles.push(newComment.id)
            let currentLens = this.codeLenses.get(newComment.id)
            if (currentLens) {
                currentLens.addRange(thread.range)
            } else {
                currentLens = await this.makeCodeLenses(newComment.id, thread.uri, this.extensionPath)
                currentLens.addRange(thread.range)
                this.codeLenses.set(newComment.id, currentLens)
            }
            currentLens.updatePendingStatus(true)
            this.currentTaskId = newComment.id
        }
        this.threads = threads
        this.thread = thread
        this.selection = await this.getSelection(isFixMode)
        void vscode.commands.executeCommand('setContext', 'cody.replied', false)
    }

    /**
     * List response from Cody as comment
     */
    public reply(text: string): void {
        if (!this.thread) {
            return
        }
        const codyReply = new FileChatMessage(
            this.markdown(text.replace(/:$/, '.')),
            vscode.CommentMode.Preview,
            { name: 'Cody', iconPath: this.codyIcon },
            false,
            this.thread,
            undefined
        )
        this.thread.comments = [...this.thread.comments, codyReply]
        this.thread.canReply = true
        // remove thread if this is a fixup
        if (this.currentTaskId) {
            this.currentTaskId = ''
            this.thread.dispose()
        }
        void vscode.commands.executeCommand('setContext', 'cody.replied', true)
    }

    /**
     * Remove a comment thread / conversation
     */
    public delete(thread: vscode.CommentThread): void {
        if (!thread) {
            return
        }
        const comments = thread?.comments as FileChatMessage[]
        comments.map(comment => {
            this.codeLenses.get(comment.id)?.remove()
            this.contentProvider.delete(comment.id)
        })
        thread.dispose()
        this.reset()
    }

    public reset(): void {
        this.selectionRange = initRange
        this.thread = null
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

    /**
     * Get current selected lines from the comment thread.
     * Add an extra line to the end line to prevent empty selection on single line selection
     */
    public async getSelection(isFixMode: boolean): Promise<ActiveTextEditorSelection | null> {
        if (!this.thread) {
            return null
        }
        const activeDocument = await vscode.workspace.openTextDocument(this.thread.uri)
        const lineLength = activeDocument.lineAt(this.thread.range.end.line).text.length
        const endPostFix = new vscode.Position(this.thread.range.end.line, lineLength)
        const endPostAsk = new vscode.Position(this.thread.range.end.line + 1, 0)
        const selectionRange = new vscode.Range(this.thread.range.start, isFixMode ? endPostFix : endPostAsk)
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

    public setReplacementRange(newRange: vscode.Range, taskID: string): void {
        if (this.thread) {
            this.thread.range = newRange
        }
        this.selectionRange = newRange
        const lens = this.codeLenses.get(taskID)
        lens?.updatePendingStatus(false, newRange)
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
        this.diffFiles = []
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
