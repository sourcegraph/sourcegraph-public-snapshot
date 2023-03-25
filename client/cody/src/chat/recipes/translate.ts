import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { languageMarkdownID, languageNames } from '@sourcegraph/cody-shared/src/chat/recipes/langs'
import { Interaction } from '@sourcegraph/cody-shared/src/chat/transcript/interaction'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { MAX_RECIPE_INPUT_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { truncateText } from '@sourcegraph/cody-shared/src/prompt/truncation'
import { getShortTimestamp } from '@sourcegraph/cody-shared/src/timestamp'

import { Editor } from '../../editor'

import { Recipe } from './recipe'

export class TranslateToLanguage implements Recipe {
    public getID(): string {
        return 'translate-to-language'
    }

    public async getInteraction(
        _humanChatInput: string,
        editor: Editor,
        _intentDetector: IntentDetector,
        _codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        const toLanguage = await editor.showQuickPick(languageNames)
        if (!toLanguage) {
            // TODO: Show the warning within the Chat UI.
            // editor.showWarningMessage('Must pick a language to translate to.')
            return null
        }

        const timestamp = getShortTimestamp()
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)

        const promptMessage = `Translate the following code into ${toLanguage}\n\`\`\`\n${truncatedSelectedText}\n\`\`\``
        const displayText = renderMarkdown(
            `Translate the following code into ${toLanguage}\n\`\`\`\n${selection.selectedText}\n\`\`\``
        )

        const markdownID = languageMarkdownID[toLanguage] || ''
        const assistantResponsePrefix = `Here is the code translated to ${toLanguage}:\n\`\`\`${markdownID}\n`

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText, timestamp },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
                displayText: '',
                timestamp,
            },
            Promise.resolve([])
        )
    }
}
