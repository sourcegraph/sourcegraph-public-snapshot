import { SourcegraphRestAPIClient } from '../../sourcegraph-api/rest'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'

interface EraserAPIResponse {
    imageUrl: string
    createEraserFileUrl: string
}

export class GenerateDiagram implements Recipe {
    public id: RecipeID = 'generate-diagram'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const { promptText, displayText } = this.getPromptText(context)

        if (!promptText) {
            await showError(context, 'Select some code to generate a diagram')
            return null
        }

        // Execute promise in the background, and send the response using the responseMultiplexer
        // Note: this relies on ChatViewProvider to call runServiceRecipe
        void this.callEraser({ promptText, restApiClient: context.restApiClient })
            .then(formatResponse)
            .then((md: string) => {
                // Pass formatted response to chat UI and end interaction
                context.responseMultiplexer.publish(md)
                context.responseMultiplexer.notifyTurnComplete()
            })
            .catch(async err => {
                // Pass the message to the Chat UI and end interaction
                // Also notify more prominently via toast
                const msg = `Error generating diagram: ${err.message}`
                context.responseMultiplexer.publish(msg)
                context.responseMultiplexer.notifyTurnComplete()
                await showError(context, msg)
            })

        return new Interaction(
            { speaker: 'human', text: '', displayText },
            {
                speaker: 'assistant',
                text: '',
                displayText: 'Generating diagram...',
            },
            Promise.resolve([]),
            []
        )
    }

    private getPromptText(context: RecipeContext): { promptText: string | undefined; displayText: string } {
        // TODO: if no selected text, grab an open file
        const selection = context.editor.getActiveTextEditorSelectionOrEntireFile()

        return { promptText: selection?.selectedText, displayText: 'Diagram my selection' }
    }

    private async callEraser({
        promptText,
        restApiClient,
        // Set background to true so that the diagram shows up no matter what theme is used
        // Most developers are going to use a dark-ish theme, so we'll default to dark mode
        theme = 'dark',
        background = true,
    }: {
        restApiClient: SourcegraphRestAPIClient
        promptText: string
        theme?: 'dark' | 'light'
        background?: boolean
    }): Promise<EraserAPIResponse> {
        // TODO: verify this path
        const path = '/services/eraser/diagram'

        const body = { text: promptText, theme, background }

        const data = await restApiClient.fetch<EraserAPIResponse>(path, {
            method: 'POST',
            body: JSON.stringify(body),
        })

        return data
    }
}

function formatResponse({ imageUrl, createEraserFileUrl }: EraserAPIResponse): string {
    return `
Here is your Eraser diagram:\n\n
![Eraser diagram](${imageUrl})\n\n
Save and edit this diagram in [Eraser](${createEraserFileUrl})`
}

async function showError(context: RecipeContext, message: string): Promise<void> {
    await context.editor.controllers?.inline.error()
    await context.editor.showWarningMessage(message)
}
