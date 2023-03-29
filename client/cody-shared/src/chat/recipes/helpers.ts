import path from 'path'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { populateCodeContextTemplate } from '../../prompt/templates'

export const MARKDOWN_FORMAT_PROMPT = 'Enclose code snippets with three backticks like so: ```.'

const EXTENSION_TO_LANGUAGE: { [key: string]: string } = {
    py: 'Python',
    rb: 'Ruby',
    md: 'Markdown',
    php: 'PHP',
    js: 'Javascript',
    ts: 'Typescript',
    jsx: 'JSX',
    tsx: 'TSX',
}

export function getNormalizedLanguageName(extension: string): string {
    if (!extension) {
        return ''
    }
    const language = EXTENSION_TO_LANGUAGE[extension]
    if (language) {
        return language
    }
    return extension.charAt(0).toUpperCase() + extension.slice(1)
}

export async function getContextMessagesFromSelection(
    selectedText: string,
    precedingText: string,
    followingText: string,
    fileName: string,
    codebaseContext: CodebaseContext
): Promise<ContextMessage[]> {
    const selectedTextContext = await codebaseContext.getContextMessages(selectedText, {
        numCodeResults: 4,
        numTextResults: 0,
    })

    return selectedTextContext.concat(
        [precedingText, followingText].flatMap(text =>
            getContextMessageWithResponse(populateCodeContextTemplate(text, fileName), fileName)
        )
    )
}

export function getFileExtension(fileName: string): string {
    return path.extname(fileName).slice(1).toLowerCase()
}
