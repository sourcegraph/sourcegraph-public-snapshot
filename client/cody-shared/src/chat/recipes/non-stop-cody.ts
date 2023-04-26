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

    constructor() {
        // TODO: Dispose the subscription. Array of disposables?
        const subscription = vscode.workspace.onDidChangeTextDocument(this.textDocumentChanged.bind(this))
        this.decoCodyContribution = vscode.window.createTextEditorDecorationType({
            backgroundColor: '#0ca67888', // oc-teal-7; TODO(dpc): Account for themes. See: light, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
            // TODO: Gutter icon w/ Cody branding could be cool
        })
        this.decoCodyContributionFade = vscode.window.createTextEditorDecorationType({
            backgroundColor: 'orange', // oc-teal-7; TODO(dpc): Account for themes. See: light, dark.
            rangeBehavior: vscode.DecorationRangeBehavior.ClosedClosed,
            // TODO: Gutter icon w/ Cody branding could be cool
        })
    }

    private textDocumentChanged(event: vscode.TextDocumentChangeEvent): void {
        let decorations = this.decorations.get(event.document.uri)
        if (!decorations) {
            return
        }
        const decorationsToDelete: TrackedDecoration[] = []
        for (const decoration of decorations) {
            for (const change of event.contentChanges) {
                const updatedRange = updateRange(decoration.range, change)
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
        // TODO: Prompt the user for additional direction.

        const deco = {
            hoverMessage: 'Edited by Cody', // TODO: Put the prompt in here
            range: vscode.window.activeTextEditor!.selection,
            // TODO: Render options
        }
        let decorations: TrackedDecoration[]
        if (this.decorations.has(vscode.window.activeTextEditor!.document.uri)) {
            decorations = this.decorations.get(vscode.window.activeTextEditor!.document.uri)!
        } else {
            decorations = []
            this.decorations.set(vscode.window.activeTextEditor!.document.uri, decorations)
        }
        decorations!.push(deco)
        vscode.window.activeTextEditor?.setDecorations(this.decoCodyContribution, decorations)

        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            await context.editor.showWarningMessage('NON STOP!!!')
            return null
        }

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selection.selectedText, quarterFileContext * 2) !== selection.selectedText) {
            await context.editor.showWarningMessage("The amount of text selected exceeds Cody's current capacity.")
            return null
        }

        context.responseMultiplexer.sub(
            'selection',
            new BufferedBotResponseSubscriber(async content => {
                if (!content) {
                    await context.editor.showWarningMessage(
                        'Cody did not suggest any replacement.\nTry starting a new conversation with Cody.'
                    )
                    return
                }
                // TODO: Reinstate this
                // await context.editor.replaceSelection(selection.fileName, selection.selectedText, content)
            })
        )

        const prompt = `This is part of the file ${
            selection.fileName
        }. The part of the file I have selected is highlighted with <selection> tags. You are helping me to work on that part.

Follow the instructions in the selected part and produce a rewritten replacement for only that part. Put the rewritten replacement inside <selection> tags.

I only want to see the code within <selection>. Do not move code from outside the selection into the selection in your reply.

It is OK to provide some commentary before you tell me the replacement <selection>. If it doesn't make sense, you do not need to provide <selection>.

\`\`\`\n${truncateTextStart(selection.precedingText, quarterFileContext)}<selection>${
            selection.selectedText
        }</selection>${truncateText(
            selection.followingText,
            quarterFileContext
        )}\n\`\`\`\n\n${context.responseMultiplexer.prompt()}`
        // TODO: Move the prompt suffix from the recipe to the chat view. It may have other subscribers.

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: prompt,
                    displayText: 'Replace the instructions in the selection.',
                },
                { speaker: 'assistant' },
                this.getContextMessages(selection.selectedText, context.codebaseContext)
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
