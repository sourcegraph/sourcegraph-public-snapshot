import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_RECIPE_INPUT_TOKENS, MAX_RECIPE_SURROUNDING_TOKENS } from '../../prompt/constants'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { renderMarkdown } from '../markdown'
import { Interaction } from '../transcript/interaction'

import {
    MARKDOWN_FORMAT_PROMPT,
    getNormalizedLanguageName,
    getFileExtension,
    getContextMessagesFromSelection,
} from './helpers'
import { Recipe } from './recipe'

export class GenerateTest implements Recipe {
    getID(): string {
        return 'generate-unit-test'
    }

    public async getInteraction(
        _humanChatInput: string,
        editor: Editor,
        _intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        const timestamp = getShortTimestamp()
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)
        const truncatedPrecedingText = truncateTextStart(selection.precedingText, MAX_RECIPE_SURROUNDING_TOKENS)
        const truncatedFollowingText = truncateText(selection.followingText, MAX_RECIPE_SURROUNDING_TOKENS)
        const extension = getFileExtension(selection.fileName)

        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptMessage = `Generate a unit test in ${languageName} for the following code:\n\`\`\`${extension}\n${truncatedSelectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
        const assistantResponsePrefix = `Here is the generated unit test:\n\`\`\`${extension}\n`

        const displayText = renderMarkdown(
            `Generate a unit test for the following code:\n\`\`\`${extension}\n${selection.selectedText}\n\`\`\``
        )

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText, timestamp },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
                displayText: '',
                timestamp,
            },
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
