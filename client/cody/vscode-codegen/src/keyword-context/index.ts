import { Message } from '../sourcegraph-api'

export interface KeywordContextFetcher {
    getContextMessages(query: string): Promise<Message[]>
}

export function getTermScore(term: string): number {
    const termLength = term.length
    return term.match(/^(?=.*[a-z])(?=.*[A-Z])/) ? 10 * termLength : termLength
}
