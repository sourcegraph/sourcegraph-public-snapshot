import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'
import { isError } from '../utils'

import { Guardrails, Attribution } from '.'

export class SourcegraphGuardrailsClient implements Guardrails {
    private clients: SourcegraphGraphQLAPIClient[]

    constructor(client: SourcegraphGraphQLAPIClient) {
        this.clients = [client]
        // We want to use dotcom since that has a much larger corpus.
        if (!client.isDotCom()) {
            // Note: this is an anonymous request. We intend on adding a
            // guardrails specific API to sourcegraph which will do
            // authenticated federation to dotcom.
            this.clients.push(
                new SourcegraphGraphQLAPIClient({
                    serverEndpoint: 'https://sourcegraph.com',
                    accessToken: null,
                    customHeaders: {},
                })
            )
        }
    }

    public async searchAttribution(snippet: string): Promise<Attribution | Error> {
        // TODO(keegancsmith) adjust implementation to respect line count thresholds
        const query = `type:file select:repo content:${goEscapeString(snippet)}`
        const results = await Promise.all(this.clients.map(client => client.searchTypeRepo(query)))

        // aggregate unique repos by name
        const seen = new Set<string>()
        const aggregate: Attribution = {
            limitHit: false,
            repositories: [],
        }

        for (const result of results) {
            if (isError(result)) {
                return result
            }
            aggregate.limitHit = aggregate.limitHit || result.limitHit
            for (const repo of result.repositories) {
                if (seen.has(repo.name)) {
                    continue
                }
                seen.add(repo.name)
                aggregate.repositories.push(repo)
            }
        }

        return aggregate
    }
}

function goEscapeString(str: string): string {
    // TODO(keegancsmith) verify correct, this is blind copy pasta from cody
    let escaped = ''
    for (const c of str) {
        switch (c) {
            case '\n':
                escaped += '\\n'
                break
            case '\t':
                escaped += '\\t'
                break
            case '\r':
                escaped += '\\r'
                break
            case '\v':
                escaped += '\\v'
                break
            case '\b':
                escaped += '\\b'
                break
            case '\f':
                escaped += '\\f'
                break
            case '\\':
                escaped += '\\\\'
                break
            case '"':
                escaped += '\\"'
                break
            default:
                escaped += c
        }
    }
    return `"${escaped}"`
}
