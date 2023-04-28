import { isDefined } from '@sourcegraph/common'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { populateCurrentEditorContextTemplate } from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { SourcegraphGraphQLAPIClient } from '../../sourcegraph-api/graphql/client'
import { SEARCH_SYMBOL } from '../../sourcegraph-api/graphql/queries'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class ChatQuestion implements Recipe {
    public id = 'chat-question'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: truncatedText, displayText: humanChatInput },
                { speaker: 'assistant' },
                this.getContextMessages(truncatedText, context.editor, context.intentDetector, context.codebaseContext)
            )
        )
    }

    private async getContextMessages(
        text: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []

        const fullConfig = {
            serverEndpoint: 'https://sourcegraph.test:3443',
            accessToken: '<ADD ACCESS TOKEN>',
            debug: false,
            customHeaders: {},
        }
        const graphqlClient = new SourcegraphGraphQLAPIClient(fullConfig)

        const symbolNames = text.split(' ').map(this.looksLikeSymbol).filter(isDefined)
        console.log({ symbolNames })

        // TODO: Add repo
        const query = `context:global repo:^github\.com/hashicorp/go-multierror$ type:symbol count:50 ${symbolNames.join(
            ' '
        )}`
        const test = await graphqlClient.fetchSourcegraphAPI(SEARCH_SYMBOL, { query })
        console.log('🚀 ~ file: chat-question.ts:55 ~ ChatQuestion ~ query:', query)
        console.log('🚀 ~ file: chat-question.ts:56 ~ ChatQuestion ~ test:', test)

        const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        if (isCodebaseContextRequired) {
            const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
                numCodeResults: 12,
                numTextResults: 3,
            })
            contextMessages.push(...codebaseContextMessages)

            contextMessages.push({
                speaker: 'human',
                fileName: test.data.search.results.results[0].file.path,
                text: `
                    Here is the path to the file ${test.data.search.results.results[0].file.path}.

                    The kind of the symbol is a ${test.data.search.results.results[0].symbols.kind}.
                    The name of the symbol is a ${test.data.search.results.results[0].symbols.name}.
                    It is located in ${test.data.search.results.results[0].symbols.url}
                    This is the content of the file ${test.data.search.results.results[0].file.content}.
                `,
            })
            contextMessages.push({
                speaker: 'assistant',
                text: 'okay',
            })
        }

        if (isCodebaseContextRequired || intentDetector.isEditorContextRequired(text)) {
            contextMessages.push(...this.getEditorContext(editor))
        }

        return contextMessages
    }

    private getEditorContext(editor: Editor): ContextMessage[] {
        const visibleContent = editor.getActiveTextEditorVisibleContent()
        if (!visibleContent) {
            return []
        }
        const truncatedContent = truncateText(visibleContent.content, MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCurrentEditorContextTemplate(truncatedContent, visibleContent.fileName),
            visibleContent.fileName
        )
    }

    private looksLikeSymbol(text: string): string | undefined {
        if (text.startsWith('`') && text.endsWith('`')) {
            // If we have `variable` trim the ticks
            return text.substring(1, text.length - 1)
        }

        let numLower = 0
        let numUpper = 0
        let numDigit = 0
        let numUnder = 0
        for (let i = 0; i < text.length; i++) {
            if (text.charAt(i) >= 'a' && text.charAt(i) <= 'z') {
                numLower++
            }
            if (text.charAt(i) >= 'A' && text.charAt(i) <= 'Z') {
                numUpper++
            }
            if (text.charAt(i) >= '0' && text.charAt(i) <= '9') {
                numDigit++
            }
            if (text.charAt(i) === '_') {
                numUnder++
            }
        }

        const hasDigits = numDigit > 0
        const isSnakeCase = numUnder > 0
        const startsWithUpper = text.charAt(0) >= 'A' && text.charAt(0) <= 'Z'
        const isCamelCase = numLower !== 0 && numUpper !== 0 && (numUpper !== 1 || !startsWithUpper)
        if (hasDigits || isSnakeCase || isCamelCase) {
            return text
        }

        return undefined
    }
}
