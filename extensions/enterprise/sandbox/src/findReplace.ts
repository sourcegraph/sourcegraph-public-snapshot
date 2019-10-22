import { Subscription, Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
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

async function rewrite(_diagnostic: any, context: FindReplaceCampaignContext): Promise<sourcegraph.WorkspaceEdit> {
    console.log({ context })
    const { data, errors } = await queryGraphQL({
        query: `
                query Comby($repositoryNames: [String!]!, $matchTemplate: String!, $rule: String, $rewriteTemplate: String!) {
                    comby(repositoryNames: $repositoryNames, matchTemplate: $matchTemplate, rule: $rule, rewriteTemplate: $rewriteTemplate) {
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
            repositoryNames: context.repositoryNames || ['github.com/sd9/guava19to21-sample'],
            matchTemplate: context.matchTemplate,
            rule: context.rule,
            rewriteTemplate: context.rewriteTemplate,
        },
    })
    if (errors && errors.length > 0) {
        throw new Error(`GraphQL response error: ${errors[0].message}`)
    }
    const canonicalURLs: string[] = data.comby.results.map(
        (r: any) => `git://${r.file.commit.repository.name}?${r.file.commit.oid}#${r.file.path}`
    )
    const docs = await Promise.all(canonicalURLs.map(url => sourcegraph.workspace.openTextDocument(new URL(url))))

    const edit = new sourcegraph.WorkspaceEdit()
    for (const [i, doc] of docs.entries()) {
        if (doc.text!.length > 15000) {
            continue // TODO!(sqs): skip too large
        }
        edit.set(new URL(doc.uri), [sourcegraph.TextEdit.patch(data.comby.results[i].rawDiff)])
    }

    return (edit as any).toJSON()
}
