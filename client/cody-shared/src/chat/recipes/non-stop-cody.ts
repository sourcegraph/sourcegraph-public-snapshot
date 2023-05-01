import * as vscode from 'vscode'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'
import { updateRange } from './tracked-range'

type TrackedDecoration = vscode.DecorationOptions

// TODO(dpc): This is similar to Cody: Fixup so if it works well, integrate them.
export class NonStopCody implements Recipe {
    public id = 'non-stop-cody'
    private decoCodyContribution: vscode.TextEditorDecorationType
    private decoCodyContributionFade: vscode.TextEditorDecorationType
    private tick = 0
    private decorations: Map<vscode.Uri, TrackedDecoration[]> = new Map()
    private comments: vscode.CommentController
    private thread?: vscode.CommentThread

    constructor() {
        // TODO: Dispose the subscription. Array of disposables?
        const subscription = vscode.workspace.onDidChangeTextDocument(this.textDocumentChanged.bind(this))
        this.decoCodyContribution = vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67888', // oc-teal-7; TODO(dpc): Account for themes. See: fhtlight, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
            // TODO: Gutter icon w/ Cody branding could be cool
        })
        this.decoCodyContributionFade = vscode.window.createTextEditorDecorationType({
            backgroundColor: 'orange', // oc-teal-7; TODO(dpc): Account for themes. See: light, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
            // TODO: Gutter icon w/ Cody branding could be cool
        })

        this.comments = vscode.comments.createCommentController('cody', 'Cody')
        this.comments.options = {
            prompt: 'Hello, world',
            placeHolder: 'Replace me',
        }
    }

    private textDocumentChanged(event: vscode.TextDocumentChangeEvent): void {
        let decorations = this.decorations.get(event.document.uri)
        if (!decorations) {
            return
        }
        const decorationsToDelete: TrackedDecoration[] = []
        for (const decoration of decorations) {
            for (const change of event.contentChanges) {
                const updatedRange = updateRange<vscode.Range, vscode.Position>(decoration.range, change)
                if (updatedRange) {
                    decoration.range = updatedRange
                } else {
                    decorationsToDelete.push(decoration)
                }
            }
            if (decorationsToDelete) {
                decorations = decorations.filter(decoration => !decorationsToDelete.includes(decoration))
                this.decorations.set(event.document.uri, decorations)
            }
        }
        this.tick++
        const oldHi = this.tick % 2 === 0 ? this.decoCodyContribution : this.decoCodyContributionFade
        const newHi = this.tick % 2 === 1 ? this.decoCodyContribution : this.decoCodyContributionFade
        // TODO: Also need to listen to the active editor change event and update highlights
        if (vscode.window.activeTextEditor?.document === event.document) {
            vscode.window.activeTextEditor.setDecorations(oldHi, [])
            vscode.window.activeTextEditor.setDecorations(newHi, decorations)
        }
    }

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const editor = vscode.window.activeTextEditor

        if (!editor) {
            await vscode.window.showErrorMessage('Open a text editor to use Cody: Fixup')
            await vscode.commands.executeCommand('cody.focus')
            return null
        }

        let resolvePrompt: (prompt: string) => void
        const promptPromise: Promise<string> = new Promise(resolve => {
            resolvePrompt = resolve
        })

        // TODO: Re-use a single QuickPick instead of creating one each time
        const quickPick = vscode.window.createQuickPick()
        quickPick.title = 'Cody: Fixup'
        quickPick.placeholder = 'Cody, you should...'
        // TODO: When Cody has diffs ready, include an item to commit the diffs first.
        quickPick.onDidAccept(() => {
            resolvePrompt(quickPick.value)
            quickPick.dispose()
        })
        quickPick.onDidHide(() => {
            // Note, this event happens after onDidAccept. In that case the
            // Promise is already resolved and we do nothing.
            resolvePrompt('')
            quickPick.dispose()
        })
        quickPick.show()

        const userPrompt = await promptPromise
        if (!userPrompt) {
            return null
        }

        // const deco = {
        //     hoverMessage: 'Edited by Cody', // TODO: Put the prompt in here
        //     range: vscode.window.activeTextEditor!.selection,
        //     // TODO: Render options
        // }
        // let decorations: TrackedDecoration[]
        // if (this.decorations.has(vscode.window.activeTextEditor!.document.uri)) {
        //     decorations = this.decorations.get(vscode.window.activeTextEditor!.document.uri)!
        // } else {
        //     decorations = []
        //     this.decorations.set(vscode.window.activeTextEditor!.document.uri, decorations)
        // }
        // decorations!.push(deco)
        // vscode.window.activeTextEditor?.setDecorations(this.decoCodyContribution, decorations)

        const selection = editor.selection

        // Drop a comment in the document.
        // TODO: Elaborate the comment UI to let people interact with queued tasks.
        const thread = this.comments.createCommentThread(editor.document.uri, selection, [
            {
                body: userPrompt,
                mode: vscode.CommentMode.Preview,
                author: { name: 'You' },
            },
        ])
        this.thread = thread
        thread.collapsibleState = vscode.CommentThreadCollapsibleState.Collapsed

        const precedingText = editor.document.getText(
            new vscode.Range(
                selection.start.translate({ lineDelta: Math.max(-50, -selection.start.line) }),
                editor.selection.start
            )
        )
        const selectedText = editor.document.getText(selection)
        const followingText = editor.document.getText(
            new vscode.Range(selection.end, selection.end.translate({ lineDelta: 50 }))
        )

        context.responseMultiplexer.sub(
            'cody-replace',
            new BufferedBotResponseSubscriber(async content => {
                if (!content) {
                    // TODO: Put a button here to restart the conversation.
                    await vscode.window.showWarningMessage(
                        'Cody did not suggest any replacement.\nTry starting a new conversation with Cody.'
                    )
                    return
                }
                // TODO: Animate diff availability
                // await context.editor.replaceSelection(selection.fileName, selection.selectedText, content)
                const success = await editor.edit(
                    edit => {
                        edit.insert(new vscode.Position(0, 0), content + '\n<FIN>\n')
                    },
                    { undoStopAfter: true, undoStopBefore: true }
                )
                await vscode.window.showInformationMessage(
                    `done, generated ${content.length} characters; edit: ${success}`
                )
            })
        )

        // TODO: Move this LLM interaction outside of the recipe, so we can queue multiple changes at once.
        // TODO: Bring back prompting which limits changes to the selection.

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selectedText, quarterFileContext * 2) !== selectedText) {
            await context.editor.showWarningMessage("The amount of text selected exceeds Cody's current capacity.")
            return null
        }

        // TODO: This hardcodes the Anthropic "Assistant:", "Human:" prompts. Need to generalize this.
        const prompt = `I need your help to improve some code. The area I need help with is highlighted with <cody-help> tags. You are helping me work on that part. Follow the instructions in the prompt attribute and produce a rewritten replacement. You should remove the <cody-help> tags from your replacement. Put the replacement in <cody-replace> tags. I need only the replacement, no other commentary about it. Do not write anything after the closing </cody-replace> tag.

Assistant: OK, I understand. I will follow the prompts to improve the code, and only reply with code in <cody-replace> tags. The last thing I write will be the closing </cody-replace> tag.

Human: Wonderful. This is part of the file warmup.js:

function greet() {
    console.log('hello, world');
}

<cody-help prompt="Document this function.">
function clipRange(start, end) {
    return (x) => {
        if (x < start) {
            return start;
        } else if (x > end) {
            return end;
        }
    }
}
<cody-help prompt="Add error checking for the parameters."></cody-help>
</cody-help>

function back() {
    window.history.go(-1);
}

Assistant: <cody-replace>
function greet() {
    console.log('hello, world');
}

// Creates a function which clips a value to a range.
function clipRange(start, end) {
    if (end < start) {
        throw new Error('invalid range: start must be at or before end');
    }
    return (x) => {
        if (x < start) {
            return start;
        } else if (x > end) {
            return end;
        }
    }
}

function back() {
    window.history.go(-1);
}
</cody-replace>

Human: Great! That is perfect. Now, this is part of the file ${
            editor.document.fileName
        }. The area I need help with is highlighted with <cody-help> tags. Again, I only need the replacement in <cody-replace> tags.

${truncateTextStart(
    precedingText,
    quarterFileContext
)}<cody-help prompt="${userPrompt}">${selectedText}</cody-help>${truncateText(
            followingText,
            quarterFileContext
        )}\n\n${context.responseMultiplexer.prompt()}`
        // TODO: Move the prompt suffix from the recipe to the chat view. It may have other subscribers.

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: prompt,
                    displayText: 'Replace the instructions in the selection.',
                },
                { speaker: 'assistant' },
                this.getContextMessages(selectedText || precedingText + followingText, context.codebaseContext)
            )
        )
    }

    private async getContextMessages(text: string, codebaseContext: CodebaseContext): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = await codebaseContext.getContextMessages(text, {
            numCodeResults: 12,
            numTextResults: 3,
        })
        return contextMessages
    }
}
