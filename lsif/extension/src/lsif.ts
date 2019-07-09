import * as path from 'path'
import * as sourcegraph from 'sourcegraph'
import queryString from 'query-string'

function repositoryFromDoc(doc: sourcegraph.TextDocument): string {
    const url = new URL(doc.uri)
    return path.join(url.hostname, url.pathname.slice(1))
}

function commitFromDoc(doc: sourcegraph.TextDocument): string {
    const url = new URL(doc.uri)
    return url.search.slice(1)
}

function pathFromDoc(doc: sourcegraph.TextDocument): string {
    const url = new URL(doc.uri)
    return url.hash.slice(1)
}

function setPath(doc: sourcegraph.TextDocument, path: string): string {
    const url = new URL(doc.uri)
    url.hash = path
    return url.href
}

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(
        sourcegraph.languages.registerHoverProvider(['*'], {
            provideHover: async (doc, params) => {
                const response = await fetch(
                    path.join(
                        sourcegraph.internal.sourcegraphURL +
                            `.api/lsif/request?${queryString.stringify({
                                repository: repositoryFromDoc(doc),
                                commit: commitFromDoc(doc),
                            })}`
                    ),
                    {
                        method: 'POST',
                        headers: { 'content-type': 'application/json' },
                        body: JSON.stringify({
                            method: 'hover',
                            params: [pathFromDoc(doc), params],
                        }),
                    }
                )
                const body = await response.json()
                if (body.error) {
                    if (body.error === 'No result found') {
                        return null
                    }
                    throw new Error(body.error)
                }
                return {
                    ...body,
                    contents: {
                        value: body.contents
                            .map((content: { language: string; value: string } | string) =>
                                typeof content === 'string'
                                    ? content
                                    : content.language
                                    ? ['```' + content.language, content.value, '```'].join('\n')
                                    : content.value
                            )
                            .join('\n'),
                        kind: sourcegraph.MarkupKind.Markdown,
                    },
                }
            },
        })
    )

    ctx.subscriptions.add(
        sourcegraph.languages.registerDefinitionProvider(['*'], {
            provideDefinition: async (doc, params) => {
                const response = await fetch(
                    path.join(
                        sourcegraph.internal.sourcegraphURL +
                            `.api/lsif/request?${queryString.stringify({
                                repository: repositoryFromDoc(doc),
                                commit: commitFromDoc(doc),
                            })}`
                    ),
                    {
                        method: 'POST',
                        headers: { 'content-type': 'application/json' },
                        body: JSON.stringify({
                            method: 'definitions',
                            params: [pathFromDoc(doc), params],
                        }),
                    }
                )
                const body = await response.json()
                if (body.error) {
                    if (body.error === 'No result found') {
                        return null
                    }
                    throw new Error(body.error)
                }
                return body.map((definition: any) => ({ ...definition, uri: setPath(doc, definition.uri) }))
            },
        })
    )
}
