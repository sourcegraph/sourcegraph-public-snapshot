import * as vscode from 'vscode'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { FileChatProvider } from '../../editor'
import { MAX_CURRENT_FILE_TOKENS, SURROUNDING_LINES } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { computeDiff } from './concurrent-editing'
import { Recipe, RecipeContext } from './recipe'
import { updateRange } from './tracked-range'

type TrackedDecoration = vscode.DecorationOptions

interface BatchState {
    editor: vscode.TextEditor
    original: string
    range: vscode.Range
}

// TODO(dpc): This is similar to Cody: Fixup so if it works well, integrate them.
export class NonStopCody implements Recipe {
    public id = 'non-stop-cody'

    // TODO: Generalize this to having multiple in-flight at once.
    private batch?: BatchState
    private trackingInProgress = false
    private comments: FileChatProvider | null = null

    private decorations: Map<vscode.Uri, TrackedDecoration[]> = new Map()
    private fadeAnimation = 0
    private fadeTimer: NodeJS.Timeout | undefined = undefined
    private fadeDecorations: vscode.TextEditorDecorationType[] = [
        vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67888', // oc-teal-7; TODO(dpc): Account for themes. See: fhtlight, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
            // TODO: Gutter icon w/ Cody branding could be cool
        }),
        vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67855', // oc-teal-7; TODO(dpc): Account for themes. See: fhtlight, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
        }),
        vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67822', // oc-teal-7; TODO(dpc): Account for themes. See: fhtlight, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
        }),
        vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67811', // oc-teal-7; TODO(dpc): Account for themes. See: fhtlight, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
        }),
    ]

    constructor() {
        // TODO: Dispose the subscription. Array of disposables?
        const subscription = vscode.workspace.onDidChangeTextDocument(this.textDocumentChanged.bind(this))
    }

    private onAnimationTick(): void {
        if (!this.trackingInProgress) {
            this.fadeTimer = undefined
            return
        }
        // TODO: Make this animation tick per document
        if (this.fadeAnimation === this.fadeDecorations.length) {
            for (const editor of vscode.window.visibleTextEditors) {
                this.decorations.delete(editor.document.uri)
                editor.setDecorations(this.fadeDecorations[this.fadeDecorations.length - 1], [])
            }
            if (this.fadeTimer) {
                clearInterval(this.fadeTimer)
            }
            this.fadeTimer = undefined
            return
        }
        for (const editor of vscode.window.visibleTextEditors) {
            editor.setDecorations(this.fadeDecorations[this.fadeAnimation], [])
            const decorations = this.decorations.get(editor.document.uri)
            if (!decorations) {
                continue
            }
            editor.setDecorations(this.fadeDecorations[this.fadeAnimation + 1], decorations)
        }
        this.fadeAnimation++
    }

    private textDocumentChanged(event: vscode.TextDocumentChangeEvent): void {
        // Stop tracking changes when no task has been initiated
        if (!this.trackingInProgress) {
            return
        }
        // TODO: Experiment with a cooldown timer which commits changes when the user is idle.
        // TODO: Generalize this to tracking multiple ranges
        if (this.batch) {
            let range: vscode.Range | null = this.batch.range
            for (const change of event.contentChanges) {
                range = updateRange<vscode.Range, vscode.Position>(range, change)
                if (!range) {
                    break
                }
            }
            if (range) {
                this.batch.range = range
            } else {
                // TODO: Handle the case where the place Cody is editing has been obliterated.
                this.batch = undefined
            }
        }

        let decorations = this.decorations.get(event.document.uri)
        if (decorations) {
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
        }
    }

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        this.comments = context.editor.fileChatProvider || null
        if (!this.comments) {
            return null
        }
        const editor = vscode.window.activeTextEditor

        if (!editor) {
            await vscode.window.showErrorMessage('Open a text editor to use Cody: Fixup')
            await vscode.commands.executeCommand('cody.focus')
            return null
        }

        if (!humanChatInput) {
            return null
        }
        this.trackingInProgress = true
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
        // const deco = {
        //     hoverMessage: 'Edited by Cody', // TODO: Put the prompt in here
        //     range: vscode.window.activeTextEditor!.selection,
        //     // TODO: Render options
        // }
        // decorations!.push(deco)
        // vscode.window.activeTextEditor?.setDecorations(this.decoCodyContribution, decorations)

        const selection = editor.selection

        const providedCodeStart = selection.start.translate({
            lineDelta: Math.max(0 - SURROUNDING_LINES, -selection.start.line),
            characterDelta: -selection.start.character,
        })
        const providedCodeEnd = editor.document.validatePosition(
            selection.end.translate({ lineDelta: SURROUNDING_LINES, characterDelta: -selection.end.character })
        )
        const precedingText = editor.document.getText(new vscode.Range(providedCodeStart, editor.selection.start))
        const selectedText = editor.document.getText(selection)
        const followingText = editor.document.getText(new vscode.Range(selection.end, providedCodeEnd))

        this.batch = {
            editor,
            original: precedingText + selectedText + followingText,
            range: new vscode.Range(providedCodeStart, providedCodeEnd),
        }

        const handleResult = this.handleResult.bind(this)
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
                // TODO: Consider handling content progressively
                await handleResult(content)
            })
        )

        // TODO: Move this LLM interaction outside of the recipe, so we can queue multiple changes at once.
        // TODO: Cody chat steals the editor focus around this point. Not very "non-stop."
        // TODO: Bring back prompting which limits changes to the selection.

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selectedText, quarterFileContext * 2) !== selectedText) {
            await context.editor.showWarningMessage("The amount of text selected exceeds Cody's current capacity.")
            return null
        }

        // Prompt for Cody
        const promptText = NonStopCody.prompt
            .replace('{responseMultiplexerPrompt', context.responseMultiplexer.prompt())
            .replace('{truncateFollowingText}', truncateText(followingText, quarterFileContext))
            .replace('{selectedText}', selectedText)
            .replace('{humanInput}', humanChatInput)
            .replace('{truncateTextStart}', truncateTextStart(precedingText, quarterFileContext))
            .replace('{fileName}', editor.document.fileName)

        // TODO: Move the prompt suffix from the recipe to the chat view. It may have other subscribers.
        const text = selectedText || precedingText + selectedText + followingText
        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: 'Update the selected code based on my instructions.',
                },
                {
                    speaker: 'assistant',
                    prefix: 'Document has been updated based on your instruction.\n',
                },
                text ? this.getContextMessages(text, context.codebaseContext) : Promise.resolve([])
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

    private async handleResult(content: string): Promise<void> {
        this.trackingInProgress = false
        // TODO: Handle multiple concurrent editors, don't use activeTextEditor here but make it part of the batch
        if (!this.batch) {
            return
        }

        const editedText = this.batch.editor.document.getText(this.batch.range)
        const diff = computeDiff(this.batch.original, content, editedText, this.batch.range.start)
        if (!diff.clean) {
            await vscode.window.showErrorMessage('Cody: Diff does not apply cleanly')
            // TODO: Handle this by scheduling another spin, invoking diff mode, etc.
            this.batch = undefined
            return
        }

        // TODO: Animate diff availability
        const success = await this.batch.editor.edit(
            editBuilder => {
                for (const edit of diff.edits) {
                    editBuilder.replace(
                        new vscode.Range(
                            new vscode.Position(edit.range.start.line, edit.range.start.character),
                            new vscode.Position(edit.range.end.line, edit.range.end.character)
                        ),
                        edit.text
                    )
                }
            },
            { undoStopAfter: true, undoStopBefore: true }
        )

        // Highlight what Cody did
        const decorations: vscode.DecorationOptions[] = []
        this.decorations.set(this.batch.editor.document.uri, decorations)
        for (const highlight of diff.highlights) {
            const deco = {
                hoverMessage: 'Edited by Cody', // TODO: Put the prompt in here
                range: new vscode.Range(
                    new vscode.Position(highlight.start.line, highlight.start.character),
                    new vscode.Position(highlight.end.line, highlight.end.character)
                ),
            }
            decorations.push(deco)
        }

        vscode.window.activeTextEditor?.setDecorations(this.fadeDecorations[0], decorations)
        this.fadeAnimation = 0

        if (typeof this.fadeTimer !== 'undefined') {
            clearInterval(this.fadeTimer)
        }

        this.fadeTimer = setInterval(this.onAnimationTick.bind(this), 1500)

        await vscode.window.showInformationMessage(
            `Cody done, generated ${content.length} characters; edit: ${success}`
        )
    }

    // TODO: This hardcodes the Anthropic "Assistant:", "Human:" prompts. Need to generalize this.
    public static readonly prompt = `
    I need your help to improve some code.
    The area I need help with is highlighted with <cody-help> tags. You are helping me work on that part.
    Follow the instructions in the prompt attribute and produce a rewritten replacement.
    You should remove the <cody-help> tags from your replacement. Put the replacement in <cody-replace> tags.
    I need only the replacement, no other commentary about it. Do not write anything after the closing </cody-replace> tag.
    If you are adding code, I need you to repeat several lines of my code before and after your new code so I understand where to insert your new code.

    Assistant: OK, I understand. I will follow the prompts to improve the code, and only reply with code in <cody-replace> tags. The last thing I write will be the closing </cody-replace> tag.

    Human: Great, thank you! This is part of the file {fileName}. The area I need help with is highlighted with <cody-help> tags.
    Again, I only need the replacement in <cody-replace> tags.

    <cody-replace>{truncateTextStart}<cody-help prompt="{humanInput}">{selectedText}</cody-help>{truncateFollowingText}</cody-replace>

    {responseMultiplexerPrompt}`
}
