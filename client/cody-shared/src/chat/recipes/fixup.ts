import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class Fixup implements Recipe {
    public getID(): string {
        return 'fixup'
    }

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        // TODO: Prompt the user for additional direction.

        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            await context.editor.showWarningMessage('Select some code to fixup.')
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
                await context.editor.replaceSelection(selection.fileName, selection.selectedText, content)
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
