import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText, isTextTruncated } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { languageMarkdownID, languageNames } from './langs'
import { Recipe, RecipeContext, RecipeID } from './recipe'

export class TranslateToLanguage implements Recipe {
    public id: RecipeID = 'translate-to-language'

    public static options = languageNames

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelectionOrEntireFile()
        if (!selection) {
            await context.editor.showWarningMessage('No code selected. Please select some code and try again.')
            return null
        }

        const toLanguage = await context.editor.showQuickPick(languageNames)
        if (!toLanguage) {
            await context.editor.showWarningMessage('No language selected. Please select a language and try again.')
            return null
        }

        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)

        if (isTextTruncated(selection.selectedText, truncatedSelectedText)) {
            await context.editor.showWarningMessage('Truncated extra long selection so output may be incomplete.')
        }

        const promptMessage = `Translate the following code into ${toLanguage}\n\`\`\`\n${truncatedSelectedText}\n\`\`\``
        const displayText = `Translate the following code into ${toLanguage}\n\`\`\`\n${selection.selectedText}\n\`\`\``

        const markdownID = languageMarkdownID[toLanguage] || ''
        const assistantResponsePrefix = `Here is the code translated to ${toLanguage}:\n\`\`\`${markdownID}\n`

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            Promise.resolve([]),
            []
        )
    }
}
