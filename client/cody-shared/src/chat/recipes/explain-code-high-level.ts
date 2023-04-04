import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_RECIPE_INPUT_TOKENS, MAX_RECIPE_SURROUNDING_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { Interaction } from '../transcript/interaction'

import { getContextMessagesFromSelection, getNormalizedLanguageName, MARKDOWN_FORMAT_PROMPT } from './helpers'
import { Recipe } from './recipe'

export class ExplainCodeHighLevel implements Recipe {
    public getID(): string {
        return 'explain-code-high-level'
    }

    public async getInteraction(
        _humanChatInput: string,
        editor: Editor,
        _intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return Promise.resolve(null)
        }

        const timestamp = getShortTimestamp()
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)
        const truncatedPrecedingText = truncateTextStart(selection.precedingText, MAX_RECIPE_SURROUNDING_TOKENS)
        const truncatedFollowingText = truncateText(selection.followingText, MAX_RECIPE_SURROUNDING_TOKENS)

        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptMessage = `Explain the following ${languageName} code at a high level. Only include details that are essential to an overal understanding of what's happening in the code.\n\`\`\`\n${truncatedSelectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
        const displayText = `Explain the following code at a high level:\n\`\`\`\n${selection.selectedText}\n\`\`\``

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText, timestamp },
            { speaker: 'assistant', text: '', displayText: '', timestamp },
            getContextMessagesFromSelection(
                truncatedSelectedText,
                truncatedPrecedingText,
                truncatedFollowingText,
                selection.fileName,
                codebaseContext
            )
        )
    }
}
