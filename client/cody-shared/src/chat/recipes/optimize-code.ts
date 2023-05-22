import { MAX_RECIPE_INPUT_TOKENS, MAX_RECIPE_SURROUNDING_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import {
    MARKDOWN_FORMAT_PROMPT,
    getNormalizedLanguageName,
    getContextMessagesFromSelection,
    getFileExtension,
} from './helpers'
import { Recipe, RecipeContext } from './recipe'

export class OptimizeCode implements Recipe {
    public id = 'optimize-code'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelectionOrEntireFile()
        if (!selection) {
            return Promise.resolve(null)
        }

        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)
        const truncatedPrecedingText = truncateTextStart(selection.precedingText, MAX_RECIPE_SURROUNDING_TOKENS)
        const truncatedFollowingText = truncateText(selection.followingText, MAX_RECIPE_SURROUNDING_TOKENS)
        const extension = getFileExtension(selection.fileName)

        const displayText = `Optimize the time and space consumption of the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``

        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptMessage = `Optimise this code in  ${languageName}. \
Start your response by telling if the code can/cannot be optimized. \
You need to suggest a list of possible optimisation in less than 50 words each,\
but if no optimisation is possible just say Code is already optimised. \
If the code is optimisable, you provide Big O time/space comparison and \
return updated code, however skip these details if any of them is not changed.\
For updated code, add inline comments about changes you made and enclose it in triple backticks. \
Output format should be: The code can/cannot be optimized. Optimisation Steps: {} Time and Space Usage: {} Updated Code: {} :\n\n\`\`\`${extension}\n${truncatedSelectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
        const assistantResponsePrefix = `This code can \n\`\`\`${extension}\n`

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            getContextMessagesFromSelection(
                truncatedSelectedText,
                truncatedPrecedingText,
                truncatedFollowingText,
                selection.fileName,
                context.codebaseContext
            )
        )
    }
}
