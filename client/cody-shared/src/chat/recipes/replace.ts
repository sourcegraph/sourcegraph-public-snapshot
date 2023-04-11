import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_CURRENT_FILE_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class Replace implements Recipe {
    public getID(): string {
        return 'replace'
    }

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        // TODO: Prompt the user for additional direction.

        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            await context.editor.showWarningMessage('Select some code to operate on.')
            return null
        }

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(selection.selectedText, quarterFileContext * 2) !== selection.selectedText) {
            await context.editor.showWarningMessage("The amount of text selected exceeds Cody's current capacity.")
            return null
        }

        const prompt = `This is part of the file ${
            selection.fileName
        }. The part of the file I have selected is highlighted with <selected> tags. You are helping me to work on that part.
Follow the instructions in the selected part and produce a rewritten replacement for only that part. Put the rewritten replacement inside <selected> tags.\n\n\`\`\`\n${truncateTextStart(
            selection.precedingText,
            quarterFileContext
        )}<selection>${selection.selectedText}</selection>${truncateText(
            selection.followingText,
            quarterFileContext
        )}\n\`\`\``

        const timestamp = getShortTimestamp()

        context.responseMultiplexer.sub(
            'selected',
            new BufferedBotResponseSubscriber(async content => {
                if (!content) {
                    await context.editor.showWarningMessage('Cody did not suggest any replacement.')
                    return
                }
                await context.editor.replaceSelection(selection.fileName, selection.selectedText, content)
            })
        )

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: prompt, displayText: '', timestamp },
                { speaker: 'assistant', text: '', displayText: '', timestamp },
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
