import { MAX_RECIPE_INPUT_TOKENS, MAX_RECIPE_SURROUNDING_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import {
    MARKDOWN_FORMAT_PROMPT,
    getNormalizedLanguageName,
    getContextMessagesFromSelection,
    getFileExtension,
} from './helpers'
import { Recipe, RecipeContext, RecipeID } from './recipe'

export class OptimizeCode implements Recipe {
    public id: RecipeID = 'optimize-code'

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
        const promptMessage = `Optimize the memory and time consumption of this code in ${languageName}.\
You first tell if the code can/cannot be optimized, then\
if the code is optimizable, suggest a numbered list of possible optimizations in less than 50 words each,\
then provide Big O time/space comparison for old and new code and finally return updated code.\
Include inline comments to explain the optimizations in updated code.\
Show the old vs. new time/space optimizations in a table.\
Don't include the input code in your response. Beautify the response for better readability.\
Response format should be: This code can/cannot be optimzed. Optimization Steps: {} Time and Space Usage: {} Updated Code: {}\
However if no optimization is possible; just say the code is already optimized. \n\n\`\`\`${extension}\n${truncatedSelectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
        const assistantResponsePrefix = ''

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
                selection,
                context.codebaseContext
            ),
            []
        )
    }
}
