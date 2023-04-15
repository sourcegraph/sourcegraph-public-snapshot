import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { languageMarkdownID, languageNames } from './langs'
import { Recipe, RecipeContext } from './recipe'

export class TranslateToLanguage implements Recipe {
    public getID(): string {
        return 'translate-to-language'
    }

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        const toLanguage = await context.editor.showQuickPick(languageNames)
        if (!toLanguage) {
            // TODO: Show the warning within the Chat UI.
            // editor.showWarningMessage('Must pick a language to translate to.')
            return null
        }

        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)

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
                displayText: '',
            },
            Promise.resolve([])
        )
    }
}
