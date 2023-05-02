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

        // Check if request comes from file chat
        const isFromFileChat = humanChatInput.startsWith('/fix')

        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            await context.editor.showWarningMessage('Select some code to fixup.')
            return null
        }
        const truncatedText = truncateText(humanChatInput.replace('/fix', ''), MAX_HUMAN_INPUT_TOKENS)
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
                await context.editor.replaceSelection(selection.fileName, selection.selectedText, this.clean(content))
            })
        )

        const instructionsInFile =
            'Follow the instructions in the selected part and produce a rewritten replacement for only that part.'
        const instructionsFromChat = 'Follow my instructions and produce a rewritten replacement for only that part.'
        const prompt = `This is part of the file ${
            selection.fileName
        }. The part of the file I have selected is highlighted with <selection> tags. You are helping me to work on that part.

${isFromFileChat ? instructionsFromChat : instructionsInFile} Put the rewritten replacement inside <selection> tags.

${isFromFileChat ? 'Here is my instructions:' + truncatedText : '\n'}

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
                    displayText: humanChatInput || 'Update the document based on my instruction.',
                },
                {
                    speaker: 'assistant',
                    prefix: 'Document has been updated based on your instruction.\n',
                    text: 'Document has been updated based on your instruction.\n',
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

    private clean(text: string): string {
        const tagsIndex = text.indexOf('tags:')
        if (tagsIndex !== -1) {
            return text.slice(tagsIndex + 6).trim()
        }
        return text.trim()
    }
}
