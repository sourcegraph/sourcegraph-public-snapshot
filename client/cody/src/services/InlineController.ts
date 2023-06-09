import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

import { logEvent } from '../event-logger'
import { CodyTaskState } from '../non-stop/utils'

import { CodeLensProvider } from './CodeLensProvider'
import { editDocByUri, getIconPath, updateRangeOnDocChange } from './InlineAssist'

const initPost = new vscode.Position(0, 0)
const initRange = new vscode.Range(initPost, initPost)

export class InlineController {
    // Controller init
    private readonly id = 'cody-inline-chat'
    private readonly label = 'Cody: File Chat'
    private readonly threadLabel = 'Ask Cody...'
    private options = {
        prompt: 'Click here to ask Cody.',
        placeHolder: 'Ask Cody a question, or start with /fix to request edits (e.g. "/fix convert tabs to spaces")',
    }
    private readonly codyIcon: vscode.Uri = getIconPath('cody', this.extensionPath)
    private readonly userIcon: vscode.Uri = getIconPath('user', this.extensionPath)
    private _disposables: vscode.Disposable[] = []
    // Constroller State
    private commentController: vscode.CommentController
    public thread: vscode.CommentThread | null = null // a thread is a comment
    private currentTaskId = ''
    // Workspace State
    private workspacePath = vscode.workspace.workspaceFolders?.[0].uri
    public selection: ActiveTextEditorSelection | null = null
    public selectionRange = initRange
    // Inline Tasks States
    public isInProgress = false
    private codeLenses: Map<string, CodeLensProvider> = new Map()

    constructor(private extensionPath: string) {
        this.commentController = vscode.comments.createCommentController(this.id, this.label)
        this.commentController.options = this.options
        // Track last selection range in valid doc before an action is called
        vscode.window.onDidChangeTextEditorSelection(e => {
            if (
                e.textEditor.document.uri.scheme !== 'file' ||
                e.textEditor.document.uri.fsPath !== this.thread?.uri.fsPath
            ) {
                return
            }
            const selection = e.selections[0]
            if (selection && !this.isInProgress && this.selectionRange.end.line - 2 !== selection.start.line) {
                const range = new vscode.Range(
                    new vscode.Position(Math.max(0, selection.start.line - 1), 0),
                    new vscode.Position(Math.max(0, selection.end.line + 2), 0)
                )
                this.selectionRange = range
            }
        })
        // Track and update line diff when a task for the current selected range is being processed (this.isInProgress)
        // This makes sure the comment range and highlights are also updated correctly
        vscode.workspace.onDidChangeTextDocument(e => {
            // don't track
            if (
                !this.isInProgress ||
                !this.selectionRange ||
                e.document.uri.scheme !== 'file' ||
                e.document.uri.fsPath !== this.thread?.uri.fsPath
            ) {
                return
            }
            for (const change of e.contentChanges) {
                this.selectionRange = updateRangeOnDocChange(this.selectionRange, change.range, change.text)
            }
        })
        this._disposables.push(
            vscode.commands.registerCommand('cody.inline.decorations.remove', id => this.removeLens(id))
        )
    }
    /**
     * Getter to return instance
     */
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
        const threads = {
            text: humanInput,
            thread: this.thread,
        }
        return threads
    }
    /**
     * List response from Human as comment
     */
    public async chat(threads: vscode.CommentReply, isFixMode: boolean = false): Promise<void> {
        this.isInProgress = true
        const humanInput = threads.text
        const thread = threads.thread
        // disable reply until the task is completed
        thread.canReply = false
        thread.label = this.threadLabel
        const comment = new Comment(humanInput, 'Me', this.userIcon, isFixMode, thread, 'loading')
        thread.comments = [...thread.comments, comment]
        await this.runFixMode(isFixMode, comment, thread)
        this.thread = thread
        this.selection = await this.makeSelection(isFixMode)
        void vscode.commands.executeCommand('setContext', 'cody.replied', false)
    }
    /**
     * List response from Cody as comment
     */
    public reply(replyText = 'There was an error.'): void {
        if (!this.thread) {
            return
        }
        const codyReply = new Comment(replyText, 'Cody', this.codyIcon, false, this.thread, undefined)
        this.thread.comments = [...this.thread.comments, codyReply]
        this.thread.canReply = true
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
    /**
     * Reset class
     */
    public reset(): void {
        this.selectionRange = initRange
        this.thread = null
    }

    public async error(): Promise<void> {
        if (!this.currentTaskId) {
            return this.reply()
        }
        await this.stopFixMode(true)
    }
    /**
     * Create code lense and initiate decorators for fix mode
     */
    private async runFixMode(isFixMode: boolean, comment: Comment, thread: vscode.CommentThread): Promise<void> {
        if (!isFixMode) {
            return
        }
        const lens = await this.makeCodeLenses(comment.id, thread.uri, this.extensionPath)
        lens.updateState(CodyTaskState.asking, thread.range)
        lens.decorator.setState(CodyTaskState.asking, thread.range)
        await lens.decorator.decorate(thread.range)
        this.codeLenses.set(comment.id, lens)
        this.currentTaskId = comment.id
    }
    /**
     * Reset the selection range once replacement started by fixup has been completed
     * Then inform the dependents (eg. Code Lenses and Decorators) about the new range
     * so that they could update accordingly
     */
    private async stopFixMode(error = false, newRange?: vscode.Range): Promise<void> {
        this.isInProgress = false
        if (!this.currentTaskId) {
            return
        }
        const range = newRange || this.selectionRange
        const status = error ? CodyTaskState.error : CodyTaskState.ready
        const lens = this.codeLenses.get(this.currentTaskId)
        lens?.updateState(status, range)
        lens?.decorator.setState(status, range)
        await lens?.decorator.decorate(range)
        if (this.thread) {
            this.thread.range = range
        }
        this.currentTaskId = ''
        logEvent('CodyVSCodeExtension:inline-assist:error')
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
        // Add space when selectedText is empty --empty selectedText could cause delayed response
        const selection = {
            fileName: vscode.workspace.asRelativePath(this.thread.uri.fsPath),
            selectedText: activeDocument.getText(selectionRange) || ' ',
            precedingText,
            followingText,
        }
        this.selectionRange = selectionRange
        this.selection = selection
        return selection
    }
    /**
     * When a comment thread is open, the Editor will be switched to the comment input editor.
     * Get the current editor using the comment thread uri instead
     */
    public async makeCodeLenses(taskID: string, uri: vscode.Uri, extPath: string): Promise<CodeLensProvider> {
        const lens = new CodeLensProvider(taskID, extPath, uri)
        const activeDocument = await vscode.workspace.openTextDocument(uri)
        await lens.provideCodeLenses(activeDocument, new vscode.CancellationTokenSource().token)
        vscode.languages.registerCodeLensProvider('*', lens)
        return lens
    }

    public removeLens(id: string): void {
        this.codeLenses.get(id)?.remove()
        this.codeLenses.delete(id)
    }
    /**
     * Do replacement in document
     */
    public async replace(fileName: string, replacement: string, original: string): Promise<void> {
        const diff = original.trim() !== replacement.trim()
        if (!this.workspacePath || !replacement.trim() || !diff) {
            await this.stopFixMode(true)
            return
        }
        // Stop tracking for file changes to perfotm replacement
        this.isInProgress = false
        const chatSelection = this.getSelectionRange()
        const documentUri = vscode.Uri.joinPath(this.workspacePath, fileName)
        // const documentUri = vscode.Uri.file(docFsPath)
        const range = new vscode.Selection(chatSelection.start, new vscode.Position(chatSelection.end.line + 1, 0))
        const newRange = await editDocByUri(documentUri, { start: range.start.line, end: range.end.line }, replacement)
        await this.stopFixMode(false, newRange)
        logEvent('CodyVSCodeExtension:inline-assist:replaced')
    }
    /**
     * Return latest selection
     */
    public getSelection(): ActiveTextEditorSelection | null {
        return this.selection
    }
    /**
     * Return latest selection range
     */
    public getSelectionRange(): vscode.Range {
        return this.selectionRange
    }
    /**
     * Dispose the disposables
     */
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
    public body: string | vscode.MarkdownString
    public mode = vscode.CommentMode.Preview
    public author: vscode.CommentAuthorInformation
    constructor(
        public input: string,
        public name: string,
        public iconPath: vscode.Uri,
        public isTask: boolean,
        public parent?: vscode.CommentThread,
        public contextValue?: string
    ) {
        const timestamp = new Date(Date.now())
        this.id = timestamp.getTime().toString()
        this.body = this.markdown(input)
        this.author = { name, iconPath }
        this.label = '#' + this.id
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
}
