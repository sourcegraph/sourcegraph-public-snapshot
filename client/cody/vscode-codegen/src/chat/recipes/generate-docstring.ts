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
    getContextMessagesFromSelection,
    getFileExtension,
} from './helpers'
import { Recipe } from './recipe'

export class GenerateDocstring implements Recipe {
    getID(): string {
        return 'generate-docstring'
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
        const promptPrefix = `Generate a comment documenting the parameters and functionality of the following ${languageName} code:`
        let additionalInstructions = `Use the ${languageName} documentation style to generate a ${languageName} comment.`
        if (extension === 'java') {
            additionalInstructions = 'Use the JavaDoc documentation style to generate a Java comment.'
        } else if (extension === 'py') {
            additionalInstructions = 'Use a Python docstring to generate a Python multi-line string.'
        }
        const promptMessage = `${promptPrefix}\n\`\`\`\n${selection.selectedText}\n\`\`\`\n Only generate the documentation, do not generate the code. ${additionalInstructions} ${MARKDOWN_FORMAT_PROMPT}`

        let docStart = ''
        if (extension === 'java' || extension.startsWith('js') || extension.startsWith('ts')) {
            docStart = '/*'
        } else if (extension === 'py') {
            docStart = '"""\n'
        } else if (extension === 'go') {
            docStart = '// '
        }

        const displayText = renderMarkdown(
            `Generate documentation for the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``
        )

        const assistantResponsePrefix = `Here is the generated documentation:\n\`\`\`${extension}\n${docStart}`
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
