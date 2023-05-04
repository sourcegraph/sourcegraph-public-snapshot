import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class Fixup implements Recipe {
    public id = 'fixup'

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

        // TODO: Move the prompt suffix from the recipe to the chat view. It may have other subscribers.
        const promptText = Fixup.prompt
            .replace('{responseMultiplexerPrompt}', context.responseMultiplexer.prompt())
            .replace('{humanInput}', truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS))
            .replace('{truncateFollowingText}', truncateText(selection.followingText, quarterFileContext))
            .replace('{selectedText}', selection.selectedText)
            .replace('{truncateTextStart}', truncateTextStart(selection.precedingText, quarterFileContext))
            .replace('{fileName}', selection.fileName)

        context.responseMultiplexer.sub(
            'selection',
            new BufferedBotResponseSubscriber(async content => {
                if (!content) {
                    await context.editor.showWarningMessage(
                        'Cody did not suggest any replacement.\nTry starting a new conversation with Cody.'
                    )
                    return
                }
                await context.editor.replaceSelection(
                    selection.fileName,
                    selection.selectedText,
                    this.sanitize(content)
                )
            })
        )

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: 'Request: ' + humanChatInput,
                },
                {
                    speaker: 'assistant',
                    prefix: 'Cody has updated your document.\n\n',
                },
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

    private sanitize(text: string): string {
        const tagsIndex = text.indexOf('tags:')
        if (tagsIndex !== -1) {
            return text.slice(tagsIndex + 6).trimEnd()
        }
        return text.trimEnd()
    }

    // Prompt Templates
    public static readonly prompt = `
    This is part of the file {fileName}. The part of the file I have selected is highlighted with <selection> tags. You are helping me to work on that part as my coding assistant.
    Follow the instructions in the selected part plus the additional instructions to produce a rewritten replacement for only the selected part.
    Put the rewritten replacement inside <selection> tags. I only want to see the code within <selection>.
    Do not move code from outside the selection into the selection in your reply.
    Do not remove code inside the <selection> tags that might be being used by the code outside the <selection> tags.
    It is OK to provide some commentary before you tell me the replacement <selection>.
    It is encouraged to explain the updated code by adding comments to the replacement <selection>.
    If it doesn't make sense, you do not need to provide <selection>.

    \`\`\`
    {truncateTextStart}<selection>{selectedText}</selection>{truncateFollowingText}
    \`\`\`

    Additional Instruction:
    - {humanInput}
    - {responseMultiplexerPrompt}
`
}
