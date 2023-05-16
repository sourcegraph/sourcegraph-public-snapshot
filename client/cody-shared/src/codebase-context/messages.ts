import { Message } from '../sourcegraph-api'

export interface ContextFile {
    fileName: string
    repoName?: string
    revision?: string
}

export interface ContextMessage extends Message {
    file?: ContextFile
}

export interface OldContextMessage extends Message {
    fileName?: string
}

export function getContextMessageWithResponse(
    text: string,
    file: ContextFile,
    response: string = 'Ok.'
): ContextMessage[] {
    return [
        { speaker: 'human', text, file },
        { speaker: 'assistant', text: response },
    ]
}
