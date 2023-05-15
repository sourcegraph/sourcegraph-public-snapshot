import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { contentSanitizer } from './helpers'
import { Recipe, RecipeContext, RecipeID } from './recipe'

export class Fixup implements Recipe {
    public id: RecipeID = 'fixup'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        // TODO: Prompt the user for additional direction.
        const selection = context.editor.getActiveTextEditorSelection() || context.editor.controller?.selection
        if (!selection) {
            await context.editor.showWarningMessage('Select some code to fixup.')
            return null
        }
        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selection.selectedText, quarterFileContext * 2) !== selection.selectedText) {
            const msg = "The amount of text selected exceeds Cody's current capacity."
            await context.editor.showWarningMessage(msg)
            return null
        }

        // Reconstruct Cody's prompt using user's context
        // Replace placeholders in reverse order to avoid collisions if a placeholder occurs in the input
        // TODO: Move prompt suffix from recipe to chat view. It has other subscribers.
        const promptText = Fixup.prompt
            .replace('{humanInput}', truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS))
            .replace('{responseMultiplexerPrompt}', context.responseMultiplexer.prompt())
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
                    contentSanitizer(content)
                )
            })
        )

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: 'Fixup request ' + humanChatInput,
                },
                {
                    speaker: 'assistant',
                    prefix: 'Check your document for updates from Cody.\n',
                },
                this.getContextMessages(selection.selectedText, context.codebaseContext)
            )
        )
    }

    // Get context from editor
    private async getContextMessages(text: string, codebaseContext: CodebaseContext): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = await codebaseContext.getContextMessages(text, {
            numCodeResults: 12,
            numTextResults: 3,
        })
        return contextMessages
    }

    // Prompt Templates
    public static readonly prompt = `
    This is part of the file {fileName}. The part of the file I have selected is highlighted with <selection> tags. You are helping me to work on that part as my coding assistant.
    Follow the instructions in the selected part plus the additional instructions to produce a rewritten replacement for only the selected part.
    Put the rewritten replacement inside <selection> tags. I only want to see the code within <selection>.
    Do not move code from outside the selection into the selection in your reply.
    Do not remove code inside the <selection> tags that might be being used by the code outside the <selection> tags.
    It is OK to provide some commentary within the replacement <selection>.
    Only return provide me the replacement <selection> and nothing else.
    If it doesn't make sense, you do not need to provide <selection>.

    \`\`\`
    {truncateTextStart}<selection>{selectedText}</selection>{truncateFollowingText}
    \`\`\`

    Additional Instruction:
    - {responseMultiplexerPrompt}
    - {humanInput}
`
}
