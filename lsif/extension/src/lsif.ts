// lsp-client currently doesn't export the conversion functions from the
// top-level.
//
// tslint:disable-next-line: no-submodule-imports
import { convertHover, convertLocations } from '@sourcegraph/lsp-client/dist/lsp-conversion'
import * as sourcegraph from 'sourcegraph'
import * as LSP from 'vscode-languageserver-types'

function repositoryFromDoc(doc: sourcegraph.TextDocument): string {
    const url = new URL(doc.uri)
    return url.hostname + url.pathname
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

async function send({
    doc,
    method,
    params,
}: {
    doc: sourcegraph.TextDocument
    method: string
    params: any[]
}): Promise<any> {
    const url = new URL('.api/lsif/request', sourcegraph.internal.sourcegraphURL)
    url.searchParams.set('repository', repositoryFromDoc(doc))
    url.searchParams.set('commit', commitFromDoc(doc))

    const response = await fetch(url.href, {
        method: 'POST',
        headers: new Headers({
            'content-type': 'application/json',
            'x-requested-with': 'Sourcegraph LSIF extension',
        }),
        body: JSON.stringify({
            method,
            params,
        }),
    })
    if (!response.ok) {
        throw new Error(`LSIF /request returned ${response.statusText}`)
    }
    return await response.json()
}

const lsifDocs = new Map<string, Promise<boolean>>()

async function hasLSIF(doc: sourcegraph.TextDocument): Promise<boolean> {
    if (lsifDocs.has(doc.uri)) {
        return lsifDocs.get(doc.uri)!
    }

    const url = new URL('.api/lsif/exists', sourcegraph.internal.sourcegraphURL)
    url.searchParams.set('repository', repositoryFromDoc(doc))
    url.searchParams.set('commit', commitFromDoc(doc))
    url.searchParams.set('file', pathFromDoc(doc))

    const hasLSIFPromise = (async () => {
        const response = await fetch(url.href, {
            method: 'POST',
            headers: new Headers({ 'x-requested-with': 'Sourcegraph LSIF extension' }),
        })
        if (!response.ok) {
            throw new Error(`LSIF /exists returned ${response.statusText}`)
        }
        return await response.json()
    })()

    lsifDocs.set(doc.uri, hasLSIFPromise)

    return hasLSIFPromise
}

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(
        sourcegraph.languages.registerHoverProvider(['*'], {
            provideHover: async (doc, pos) => {
                if (!(await hasLSIF(doc))) {
                    return null
                }
                const hover: LSP.Hover | null = await send({ doc, method: 'hover', params: [pathFromDoc(doc), pos] })
                if (!hover) {
                    return null
                }
                return convertHover(sourcegraph, hover)
            },
        })
    )

    ctx.subscriptions.add(
        sourcegraph.languages.registerDefinitionProvider(['*'], {
            provideDefinition: async (doc, pos) => {
                if (!(await hasLSIF(doc))) {
                    return null
                }
                const body: LSP.Location | LSP.Location[] | null = await send({ doc, method: 'definitions', params: [pathFromDoc(doc), pos] })
                if (!body) {
                    return null
                }
                const locations = Array.isArray(body) ? body : [body]
                return convertLocations(sourcegraph, locations.map((definition: LSP.Location) => ({
                    ...definition,
                    uri: setPath(doc, definition.uri),
                })))
            },
        })
    )

    ctx.subscriptions.add(
        sourcegraph.languages.registerReferenceProvider(['*'], {
            provideReferences: async (doc, pos) => {
                if (!(await hasLSIF(doc))) {
                    return null
                }
                const locations: LSP.Location[] | null = await send({ doc, method: 'references', params: [pathFromDoc(doc), pos] })
                if (!locations) {
                    return null
                }
                return convertLocations(sourcegraph, locations.map((reference: LSP.Location) => ({
                    ...reference,
                    uri: setPath(doc, reference.uri),
                })))
            },
        })
    )
}
