import * as sourcegraph from 'sourcegraph'
import { Subscription, Unsubscribable } from 'rxjs'
import { queryGraphQL } from './util'

export const FIND_REPLACE_REWRITE_COMMAND = 'findReplace.rewrite'

export interface FindReplaceCampaignContext {
    matchTemplate: string
    rule: string | undefined
    rewriteTemplate: string
}

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(sourcegraph.commands.registerCommand(FIND_REPLACE_REWRITE_COMMAND, rewrite))
    return subscriptions
}

async function rewrite(context: FindReplaceCampaignContext): Promise<sourcegraph.WorkspaceEdit> {
    const { data, errors } = await queryGraphQL({
        query: `
                query Comby($matchTemplate: String!, $rule: String, $rewriteTemplate: String!) {
                    comby(matchTemplate: $matchTemplate, rule: $rule, rewriteTemplate: $rewriteTemplate) {
                        results {
                            file {
                                path
                                commit {
                                    oid
                                    repository {
                                        name
                                    }
                                }
                            }
                            rawDiff
                        }
                    }
                }
            `,
        vars: {
            matchTemplate: context.matchTemplate,
            rule: context.rule,
            rewriteTemplate: context.rewriteTemplate,
        },
    })
    if (errors && errors.length > 0) {
        throw new Error(`GraphQL response error: ${errors[0].message}`)
    }
    const canonicalURLs: string[] = data.comby.results.map(
        r => `git://${r.file.commit.repository.name}?${r.file.commit.oid}#${r.file.path}`
    )
    const docs = await Promise.all(canonicalURLs.map(async url => sourcegraph.workspace.openTextDocument(new URL(url))))

    const edit = new sourcegraph.WorkspaceEdit()
    for (const doc of docs) {
        if (doc.text.length > 15000) {
            continue // TODO!(sqs): skip too large
        }

        // TODO!(sqs): actually implement comby by hitting the api or something
        let i = 0
        while (i !== -1 && i < doc.text.length) {
            i = doc.text.indexOf(context.matchTemplate, i)
            if (i !== -1) {
                const start = doc.positionAt(i)
                const end = doc.positionAt(i + context.matchTemplate.length)
                edit.replace(new URL(doc.uri), new sourcegraph.Range(start, end), context.rewriteTemplate)
                i += context.matchTemplate.length
            }
        }
    }

    return edit.toJSON()
}
