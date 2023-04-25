import { CodebaseContext } from '../../codebase-context'
import { MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { getFileExtension } from './helpers'
import { Recipe, RecipeContext } from './recipe'

export class FuzzySearch implements Recipe {
    public id = 'fuzzy-search'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const truncatedText = truncateText(
            humanChatInput.replace('/search ', '').replace('/s ', ''),
            MAX_HUMAN_INPUT_TOKENS
        )

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: '',
                    displayText: humanChatInput,
                },
                {
                    speaker: 'assistant',
                    text: '',
                    displayText: await this.displaySearchResults(truncatedText, context.codebaseContext),
                },
                new Promise(resolve => resolve([]))
            )
        )
    }

    private async displaySearchResults(text: string, codebaseContext: CodebaseContext): Promise<string> {
        const resultContext = await codebaseContext.getSearchResults(text, {
            numCodeResults: 12,
            numTextResults: 3,
        })
        const rootPath = resultContext.endpoint

        let snippets = `Here are the code snippets for: ${text}\n\n`
        for (const file of resultContext.results) {
            const fileContent = this.sanitizeContent(file.content)
            const extension = getFileExtension(file.fileName)
            const uri = new URL(`/search?q=context:global+file:${file.fileName}`, rootPath).href
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
