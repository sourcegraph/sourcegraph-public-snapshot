import path from 'path'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { Interaction } from '../transcript/interaction'

import { Recipe } from './recipe'

export class ChatQuestion implements Recipe {
    public getID(): string {
        return 'chat-question'
    }

    public async getInteraction(
        humanChatInput: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const timestamp = getShortTimestamp()
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: truncatedText, displayText: humanChatInput, timestamp },
                { speaker: 'assistant', text: '', displayText: '', timestamp },
                this.getContextMessages(truncatedText, editor, intentDetector, codebaseContext)
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

        const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        if (isCodebaseContextRequired) {
            // Request 16 context files in total. That amounts to roughly 16 * 256 = 4096 tokens used for context.
            // That leaves us ~3000 tokens to include the chat history.
            const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
                numCodeResults: 13,
                numTextResults: 3,
            })
            contextMessages.push(...codebaseContextMessages)
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
            populateCurrentEditorCodeContextTemplate(truncatedContent, visibleContent.fileName),
            visibleContent.fileName,
            `You currently have \`${visibleContent.fileName}\` open in your editor, and I can answer questions about that file's contents.`
        )
    }
}

const CURRENT_EDITOR_CODE_TEMPLATE = `I have the \`{filePath}\` file opened in my editor. You are able to answer questions about \`{filePath}\`. The following code snippet is from the currently open file in my editor \`{filePath}\`:
\`\`\`{language}
{text}
\`\`\``

function populateCurrentEditorCodeContextTemplate(code: string, filePath: string): string {
    const language = path.extname(filePath).slice(1)
    return CURRENT_EDITOR_CODE_TEMPLATE.replace(/{filePath}/g, filePath)
        .replace('{language}', language)
        .replace('{text}', code)
}
