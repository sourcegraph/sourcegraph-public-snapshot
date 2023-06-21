import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'

// TODO: Disconnect recipe from chat
export class NonStop implements Recipe {
    public id: RecipeID = 'non-stop'

    public async getInteraction(taskId: string, context: RecipeContext): Promise<Interaction | null> {
        const controllers = context.editor.controllers
        if (!controllers) {
            return null
        }
        const taskParameters = await controllers.fixups.getTaskRecipeData(taskId)
        if (!taskParameters) {
            // Nothing to do.
            return null
        }
        const { instruction, fileName, precedingText, selectedText, followingText } = taskParameters

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selectedText, quarterFileContext * 2) !== selectedText) {
            const msg = "The amount of text selected exceeds Cody's current capacity."
            await context.editor.showWarningMessage(msg)
            // TODO: Communicate this error back to the FixupController
            return null
        }

        // Reconstruct Cody's prompt using user's context
        // Replace placeholders in reverse order to avoid collisions if a placeholder occurs in the input
        const promptText = NonStop.prompt
            .replace('{humanInput}', truncateText(instruction, MAX_HUMAN_INPUT_TOKENS))
            .replace('{responseMultiplexerPrompt}', context.responseMultiplexer.prompt())
            .replace('{truncateFollowingText}', truncateText(followingText, quarterFileContext))
            .replace('{selectedText}', selectedText)
            .replace('{truncateTextStart}', truncateTextStart(precedingText, quarterFileContext))
            .replace('{fileName}', fileName)

        let text = ''

        context.responseMultiplexer.sub('selection', {
            onResponse: async (content: string) => {
                text += content
                await context.editor.didReceiveFixupText(taskId, text, 'streaming')
            },
            onTurnComplete: async () => {
                await context.editor.didReceiveFixupText(taskId, text, 'complete')
            },
        })

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: 'Cody Fixups: ' + instruction,
                },
                {
                    speaker: 'assistant',
                    prefix: 'Check your document for updates from Cody.',
                },
                this.getContextMessages(selectedText, context.codebaseContext),
                []
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
    This is part of the file {fileName}. The part of the file I have selected is enclosed with the <selection> tags. You are helping me to work on that part as my coding assistant.
    Follow the instructions in the selected part along with the additional instruction provide below to produce a rewritten replacement for only the selected part.
    Put the rewritten replacement inside <selection> tags. I only want to see the code within <selection>.
    Do not move code from outside the selection into the selection in your reply.
    Do not remove code inside the <selection> tags that might be being used by the code outside the <selection> tags.
    Do not enclose replacement code with tags other than the <selection> tags.
    Do not enclose your answer with any markdown.
    Only return provide me the replacement <selection> and nothing else.
    If it doesn't make sense, you do not need to provide <selection>.

    \`\`\`
    {truncateTextStart}<selection>{selectedText}</selection>{truncateFollowingText}
    \`\`\`

    Additional Instruction:
    - {humanInput}
    - {responseMultiplexerPrompt}
`
}
