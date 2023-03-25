import { CHARS_PER_TOKEN } from '@sourcegraph/cody-shared/src/prompt/constants'

export function truncateText(text: string, maxTokens: number): string {
    const maxLength = maxTokens * CHARS_PER_TOKEN
    return text.length <= maxLength ? text : text.slice(0, maxLength)
}

export function truncateTextStart(text: string, maxTokens: number): string {
    const maxLength = maxTokens * CHARS_PER_TOKEN
    return text.length <= maxLength ? text : text.slice(-maxLength - 1)
}
