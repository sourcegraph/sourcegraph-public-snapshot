import path from 'path'

import { Message } from '@sourcegraph/cody-common'

import { Editor } from '../../editor'

import { languageMarkdownID, languageNames } from './langs'
import { Recipe, RecipePrompt } from './recipe'

export class TranslateToLanguage implements Recipe {
    getID(): string {
        return 'translateToLanguage'
    }

    async getPrompt(maxTokens: number, editor: Editor): Promise<RecipePrompt | null> {
        // Inputs
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        const toLanguage = await editor.showQuickPick(languageNames)
        if (!toLanguage) {
            void editor.showWarningMessage('Must pick a language to translate to')
            return null
        }

        // Context messages
        const contextMessages: Message[] = []

        // Get query message
        const promptMessage: Message = {
            speaker: 'you',
            text: `Translate the following code into ${toLanguage}\n\`\`\`\n${selection.selectedText}\n\`\`\``,
        }

        // Response prefix
        const markdownID = languageMarkdownID[toLanguage] || ''
        const botResponsePrefix = `Here is the code translated to ${toLanguage}:\n\`\`\`${markdownID}\n`

        return {
            displayText: promptMessage.text,
            promptMessage,
            botResponsePrefix,
            contextMessages,
        }
    }
}

export function getFileExtension(fileName: string): string {
    return path.extname(fileName).slice(1).toLowerCase()
}
