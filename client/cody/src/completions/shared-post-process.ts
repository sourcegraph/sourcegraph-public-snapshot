import { Completion } from '.'
import { truncateMultilineCompletion } from './multiline'
import { trimUntilSuffix } from './text-processing'

/**
 * This function implements post-processing logic that is applied regardless of
 * which provider is chosen.
 */
export function sharedPostProcess({
    prefix,
    suffix,
    languageId,
    multiline,
    completion,
}: {
    prefix: string
    suffix: string
    languageId: string
    multiline: boolean
    completion: Completion
}): Completion {
    let content = completion.content

    if (multiline) {
        content = truncateMultilineCompletion(content, prefix, suffix, languageId)
    }
    content = trimUntilSuffix(content, prefix, suffix)

    return {
        ...completion,
        content: content.trimEnd(),
    }
}
