import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

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
        this.id = this.timestamp.toUTCString()
    }
}

export class FileChatProvider {
    private readonly id = 'cody-file-chat'
    private readonly label = 'Cody: File Chat'
    private readonly threadLabel = 'Ask Cody...'
    private options = {
        prompt: 'Click here to ask Cody for help.',
        placeHolder:
            'Ask Cody a question, or start with /fix to have it perform edits. e.g. “How could you rewrite this in less lines?” or “/fix make the logo bigger”.',
    }
    public diffFiles: string[] = []

    private readonly codyIcon: vscode.Uri = this.getIconPath('cody')
    private readonly userIcon: vscode.Uri = this.getIconPath('user')

    private commentController: vscode.CommentController
    public threads: vscode.CommentReply | null = null
    public thread: vscode.CommentThread | null = null

    public editor: vscode.TextEditor | null = null
    public selection: ActiveTextEditorSelection | null = null
    public selectionRange: vscode.Range | null = null

    // Status trackers
    public addedLines = 0
    public isInProgress = false

    private _disposables: vscode.Disposable[] = []

    constructor(private extensionPath: string, public contentProvider: CodyContentProvider) {
        this.commentController = vscode.comments.createCommentController(this.id, this.label)
        this.commentController.options = this.options
        this.commentController.commentingRangeProvider = {
            provideCommentingRanges: (document: vscode.TextDocument) => {
                const lineCount = document.lineCount
                return [new vscode.Range(0, 0, lineCount - 1, 0)]
            },
        }
        // Track and update line of changes when the task for the current selected range is being processed
        vscode.workspace.onDidChangeTextDocument(e => {
            if (!this.isInProgress || !this.selectionRange || this.editor?.document.uri.scheme !== 'file') {
                return
            }
            for (const change of e.contentChanges) {
                if (this.selectionRange.end.line < change.range.start.line) {
                    return
                }
                let addedLines = 0
                if (change.text.includes('\n')) {
                    addedLines = change.text.split('\n').length - 1
                } else if (change.range.end.line - change.range.start.line > 0) {
                    addedLines -= change.range.end.line - change.range.start.line
                }
                this.selectionRange = new vscode.Range(
                    new vscode.Position(this.selectionRange.start.line + addedLines, 0),
                    new vscode.Position(this.selectionRange.end.line + addedLines, 0)
                )
            }
        })
    }

    public get(): vscode.CommentController {
        return this.commentController
    }

    public newThreads(humanInput: string): vscode.CommentReply | null {
        const editor = vscode.window.activeTextEditor
        if (!editor || !humanInput) {
            return null
        }
        this.thread = this.commentController.createCommentThread(editor?.document.uri, editor.selection, [])
        const threads: vscode.CommentReply = {
            text: humanInput,
            thread: this.thread,
        }
        this.threads = threads
        this.editor = editor
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
        newComment.label = newComment.id
        thread.comments = [...thread.comments, newComment]
        // disable reply until the task is completed
        thread.canReply = false
        this.threads = threads
        this.thread = thread
        this.selection = await this.getSelection(isFixMode)

        // store current working docs for diffs after fixup
        if (isFixMode) {
            await this.contentProvider.set(this.thread.uri, newComment.id)
            this.diffFiles.push(newComment.id)
        }

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
        this.thread.collapsibleState = vscode.CommentThreadCollapsibleState.Collapsed
        void vscode.commands.executeCommand('setContext', 'cody.replied', true)
    }
    /**
     * Remove comment thread / conversation
     */
    public delete(thread: vscode.CommentThread): void {
        this.contentProvider.deleteByThread(thread.comments as FileChatMessage[])
        thread.dispose()
        this.thread?.dispose()
        this.reset()
    }

    public reset(): void {
        this.selectionRange = null
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

    public setNewRange(newRange: vscode.Range): void {
        this.selectionRange = newRange
        if (this.thread) {
            this.thread.range = newRange
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
        this.diffFiles = []
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
