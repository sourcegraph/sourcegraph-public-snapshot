import { CHARS_PER_TOKEN, MAX_AVAILABLE_PROMPT_LENGTH } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { Interaction } from '../transcript/interaction'

import { getNormalizedLanguageName } from './helpers'
import { Recipe, RecipeContext } from './recipe'

export class FindCodeSmells implements Recipe {
    public getID(): string {
        return 'find-code-smells'
    }

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            return Promise.resolve(null)
        }

        const timestamp = getShortTimestamp()

        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptPrefix = `Find code smells, potential bugs, and unhandled errors in my ${languageName} code:`
        const promptSuffix = `List maximum five of them as a list (if you have more in mind, mention that these are the top five), with a short context, reasoning, and suggestion on each.
If you have no ideas because the code looks fine, feel free to say that it already looks fine.`

        // Use the whole context window for the prompt because we're attaching no files
        const maxTokenCount =
            MAX_AVAILABLE_PROMPT_LENGTH - (promptPrefix.length + promptSuffix.length) / CHARS_PER_TOKEN
        const truncatedSelectedText = truncateText(selection.selectedText, maxTokenCount)
        const promptMessage = `${promptPrefix}\n\n\`\`\`\n${truncatedSelectedText}\n\`\`\`\n\n${promptSuffix}`

        const displayText = `Find code smells in the following code: \n\`\`\`\n${selection.selectedText}\n\`\`\``

        const assistantResponsePrefix = ''
        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText, timestamp },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
                displayText: '',
                timestamp,
            },
            new Promise(resolve => resolve([]))
        )
    }
}
