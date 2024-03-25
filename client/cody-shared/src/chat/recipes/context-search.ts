import { type URI, Utils } from 'vscode-uri'

import type { CodebaseContext } from '../../codebase-context'
import { MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { getFileExtension, numResults } from './helpers'
import type { Recipe, RecipeContext, RecipeID } from './recipe'

/*
This class implements the context-search recipe.

Parameters:
- humanChatInput: The input from the human. If empty, a prompt will be shown to enter a search query.
- context: The recipe context.

Functionality:
- Gets a search query from the human input or a prompt.
- Truncates the query to MAX_HUMAN_INPUT_TOKENS.
- Searches the vector database for code and text results matching the query.
- If codebase is not embedded or if keyword context is selected, get local keyword context instead
- Returns up to 12 code results and 3 text results.
- Generates a markdown string displaying the results with file names linking to the search page for that file.
- Sanitizes the content by removing newlines, tabs and backticks before displaying.
*/

export class ContextSearch implements Recipe {
    public id: RecipeID = 'context-search'
    public title = 'Codebase Context Search'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const query =
            humanChatInput?.replace(/^\/s(earch)?(\s)?/i, '') ||
            (await context.editor.showInputBox('Enter your search query here...')) ||
            ''
        if (!query) {
            return null
        }
        const truncatedText = truncateText(query.replace('/search ', '').replace('/s ', ''), MAX_HUMAN_INPUT_TOKENS)
        const workspaceRootUri = context.editor.getWorkspaceRootUri()
        return new Interaction(
            {
                speaker: 'human',
                text: '',
                displayText: query,
            },
            {
                speaker: 'assistant',
                text: '',
                displayText: await this.displaySearchResults(truncatedText, context.codebaseContext, workspaceRootUri),
            },
            new Promise(resolve => resolve([])),
            []
        )
    }

    private async displaySearchResults(
        text: string,
        codebaseContext: CodebaseContext,
        workspaceRootUri: URI | null
    ): Promise<string> {
        const resultContext = await codebaseContext.getSearchResults(text, numResults)
        const endpointUri = resultContext.endpoint

        let snippets = `Here are the code snippets for: ${text}\n\n`
        for (const file of resultContext.results) {
            const fileContent = this.sanitizeContent(file.content)
            const extension = getFileExtension(file.fileName)
            const ignoreFilesExtension = /^(md|txt)$/
            if (extension.match(ignoreFilesExtension)) {
                continue
            }
            let uri: string = new URL(`/search?q=context:global+file:${file.fileName}`, endpointUri).href

            if (workspaceRootUri) {
                const fileUri = Utils.joinPath(workspaceRootUri, file.fileName)
                // TODO: make this open non-file: URIs
                uri = `vscode://file${fileUri.fsPath}`
            }

            snippets +=
                fileContent && fileContent.length > 5
                    ? `File Name: [_${file.fileName}_](${uri})\n\`\`\`${extension}\n${fileContent}\n\`\`\`\n\n`
                    : ''
        }

        return snippets
    }

    private sanitizeContent(content: string): string {
        return content.replace('\n', '').replace('\t', '').replace('`', '').trim()
    }
}
