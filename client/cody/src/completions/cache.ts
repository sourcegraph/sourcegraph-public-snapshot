import { LRUCache } from 'lru-cache'

import { Completion } from '.'

export class CompletionsCache {
    private cache = new LRUCache<string, Completion[]>({
        max: 500, // Maximum input prefixes in the cache.
    })

    // TODO: The caching strategy only takes the file content prefix into
    // account. We need to add additional information like file path or suffix
    // to make sure the cache does not return undesired results for other files
    // in the same project.
    public get(prefix: string): Completion[] | undefined {
        const results = this.cache.get(prefix)
        if (results) {
            return results.map(result => {
                if (prefix.length === result.prefix.length) {
                    return result
                }

                // Cached results can be created by appending characters from a
                // recommendation from a smaller input prompt. If that's the
                // case, we need to slightly change the content and remove
                // characters that are now part of the prefix.
                const sliceChars = prefix.length - result.prefix.length
                return {
                    ...result,
                    prefix,
                    content: result.content.slice(sliceChars),
                }
            })
        }
        return undefined
    }

    public add(completions: Completion[]): void {
        for (const completion of completions) {
            // Cache the exact prefix first and then append characters from the
            // completion one after the other until the first line is exceeded.
            //
            // If the completion starts with a `\n`, this logic will append the
            // second line instead.
            let maxCharsAppended = completion.content.indexOf('\n', completion.content.at(0) === '\n' ? 1 : 0)
            if (maxCharsAppended === -1) {
                maxCharsAppended = completion.content.length
            }

            for (let i = 0; i <= maxCharsAppended; i++) {
                const key = completion.prefix + completion.content.slice(0, i)
                if (!this.cache.has(key)) {
                    this.cache.set(key, [completion])
                } else {
                    const existingCompletions = this.cache.get(key)!
                    existingCompletions.push(completion)
                }
            }
        }
    }
}
