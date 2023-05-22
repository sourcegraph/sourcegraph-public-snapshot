import { parseMarkdown } from '../chat/markdown'
import { isError } from '../utils'

export interface RepositoryAttribution {
    name: string
}

export interface Guardrails {
    searchAttribution(snippet: string): Promise<RepositoryAttribution[] | Error>
}

/**
 * Returns markdown text with attribution information added in.
 *
 * @param guardrails client to use to lookup if a snippet of codes attributions
 * @param text markdown text
 */
export async function annotateAttribution(guardrails: Guardrails, text: string): Promise<string> {
    const tokens = parseMarkdown(text)
    const parts = await Promise.all(
        tokens.map(async token => {
            if (token.type !== 'code') {
                return token.raw
            }

            const msg = await guardrails.searchAttribution(token.text).then(result => {
                if (isError(result)) {
                    return `guardrails attribution search failed: ${result.message}`
                }

                const count = result.length
                if (count === 0) {
                    return 'no matching repositories found'
                }

                const summary = result.slice(0, count < 5 ? count : 5).map(repo => repo.name)
                if (count > 5) {
                    summary.push('...')
                }

                return `found ${count} matching repositories ${summary.join(', ')}`
            })

            // TODO(keegancsmith) escape msg?
            return `${token.raw}\n<div title="guardrails">🛡️ ${msg}</div>`
        })
    )
    return parts.join('')
}
