import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { contentSanitizer } from './helpers'
import { Recipe, RecipeContext, RecipeID } from './recipe'

export class NonStop implements Recipe {
    public id: RecipeID = 'non-stop'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const controllers = context.editor.controllers
        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection || !controllers) {
            await context.editor.showWarningMessage('Select some code to fixup.')
            return null
        }

        const humanInput =
            humanChatInput ||
            (await context.editor.showInputBox('Cody: Add instruction for your Non-Stop Fixup request.'))
        if (!humanInput) {
            await context.editor.showWarningMessage('Missing instruction for Non-Stop Fixup request.')
            return null
        }

        // Create a id using current data and use it as the key for response multiplexer
        const taskID = Date.now().toString(36).replace(/\d+/g, '')
        controllers.task.newTask(taskID, humanInput, selection, context.editor.getWorkspaceRootPath() || '')

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selection.selectedText, quarterFileContext * 2) !== selection.selectedText) {
            const msg = "The amount of text selected exceeds Cody's current capacity."
            await context.editor.showWarningMessage(msg)
            return null
        }

        // Reconstruct Cody's prompt using user's context
        // Replace placeholders in reverse order to avoid collisions if a placeholder occurs in the input
        const promptText = NonStop.prompt
            .replace('{humanInput}', truncateText(humanInput, MAX_HUMAN_INPUT_TOKENS))
            .replace('{responseMultiplexerPrompt}', context.responseMultiplexer.prompt())
            .replace('{truncateFollowingText}', truncateText(selection.followingText, quarterFileContext))
            .replace('{selectedText}', selection.selectedText)
            .replace('{truncateTextStart}', truncateTextStart(selection.precedingText, quarterFileContext))
            .replace('{fileName}', selection.fileName)

        context.responseMultiplexer.sub(
            'selection',
            new BufferedBotResponseSubscriber(async content => {
                await controllers.task.stopTask(taskID, contentSanitizer(content || ''))
                if (!content) {
                    await context.editor.showWarningMessage('Cody did not suggest any replacement.')
                    return
                }
            })
        )
        console.log(promptText, selection)
        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: 'Non-stop Cody: ' + humanInput,
                },
                {
                    speaker: 'assistant',
                    prefix: 'Check your document for updates from Cody for task #' + taskID,
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
    - {responseMultiplexerPrompt}
    - {humanInput}
`
}
