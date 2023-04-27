import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class CustomSelectionRecipe implements Recipe {
    public id = 'custom'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelectionOrEntireFile()
        if (!selection || selection.selectedText.length === 0) {
            context.editor.showWarningMessage('No text selected')
            return null
        }

        const input = await context.editor.showInputBox({
            prompt: 'What do you want to do with the selection?',
        })

        const promptMessage = `${input}\n\`\`\`\n${selection.selectedText}\n\`\`\`\n`

        return new Interaction(
            { speaker: 'human', displayText: promptMessage, text: promptMessage },
            { speaker: 'assistant' },
            Promise.resolve([])
        )
    }
}
