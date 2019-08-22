import * as lsp from 'vscode-languageserver'
import * as semver from 'semver'
import * as tmp from 'tmp-promise'
import RelateUrl from 'relateurl'
import {
    Document,
    Edge,
    EdgeLabels,
    ElementTypes,
    HoverResult,
    Range,
    Vertex,
    VertexLabels,
    DefinitionResult,
    Moniker,
    PackageInformation,
    ReferenceResult,
    ResultSet,
} from 'lsif-protocol'
import { MetaData } from './ms/server/protocol.compress'
import { fs, child_process } from 'mz'

interface HandlerMap {
    [K: string]: (element: any) => void
}

/**
 * Inserts LSIF dump data into Dgraph.
 */
export class DgraphImporter {
    private rootInfo: { rootUri: URL; repoUidRef: string; commitUidRef: string } | undefined
    private rdfStream!: fs.WriteStream

    constructor(private repository: string, private commit: string) {
        this.initHandlers()
    }

    /**
     * Create a mutation for each vertex and edge in the given list and apply
     * it to a transaction created in the constructor. This method will ensure
     * cleanup of the transaction.
     */
    public async import(items: (Vertex | Edge)[]): Promise<void> {
        // Create temp file to receive the request body
        const tempPath = (await tmp.file()).path
        console.log('tempPath', tempPath)

        return new Promise((resolve, reject) => {
            try {
                this.rdfStream = fs.createWriteStream(tempPath)

                console.log('about to generate rdf')
                console.time('generating rdf')

                for (const element of items) {
                    this.handleElement(element)
                }

                console.timeEnd('generating rdf')

                this.rdfStream.on('close', () => {
                    console.log('about to run live loader')
                    console.time('running live loader')
                    const cmd = `dgraph live -C -z localhost:5080 -c 1 -r ${tempPath}`
                    console.log(cmd)
                    child_process
                        .exec(cmd)
                        .then(([out, error]) => {
                            console.log('OUT=====')
                            console.log(out)
                            console.log('ERR=====')
                            console.log(error)

                            console.timeEnd('running live loader')
                        })
                        .then(resolve)
                        .catch(e => {
                            console.log('FAILURE')
                            console.timeEnd('running live loader')
                            reject(e)
                        })
                })
            } catch (e) {
                reject(e)
            } finally {
                this.rdfStream.end()
            }
        })
    }

    private vertexHandlerMap: HandlerMap = {}

    private initHandlers(): void {
        // Register vertex handlers
        this.vertexHandlerMap[VertexLabels.definitionResult] = e => this.handleDefinitionResult(e)
        this.vertexHandlerMap[VertexLabels.document] = e => this.handleDocument(e)
        this.vertexHandlerMap[VertexLabels.hoverResult] = e => this.handleHover(e)
        this.vertexHandlerMap[VertexLabels.metaData] = e => this.handleMetaData(e)
        this.vertexHandlerMap[VertexLabels.moniker] = e => this.handleMoniker(e)
        this.vertexHandlerMap[VertexLabels.packageInformation] = e => this.handlePackageInformation(e)
        this.vertexHandlerMap[VertexLabels.range] = e => this.handleRange(e)
        this.vertexHandlerMap[VertexLabels.referenceResult] = e => this.handleReferenceResult(e)
        this.vertexHandlerMap[VertexLabels.resultSet] = e => this.handleResultSet(e)
    }

    private handleElement(element: Vertex | Edge): void {
        if (element.type === ElementTypes.vertex) {
            const handler = this.vertexHandlerMap[element.label]
            if (handler) {
                handler(element)
            }
        } else {
            this.handleEdge(element)
        }
    }

    private handleDefinitionResult(vertex: DefinitionResult): void {
        // FIXME - no properties
    }

    private handleDocument(vertex: Document): void {
        if (!this.rootInfo) {
            throw new Error('No root info')
        }

        const id = uidRef(vertex.id)
        const path = decodeURIComponent(relativeUrl(this.rootInfo.rootUri, new URL(vertex.uri)))
        this.rdfStream.write(`${this.rootInfo.commitUidRef} <${EdgeLabels.contains}> ${id} .\n`)
        this.rdfStream.write(`${id} <Document.path> ${makeValue(path)} .\n`)
        this.rdfStream.write(`${id} <Document.label> "document" .\n`)
    }

    private handleHover(vertex: HoverResult): void {
        const id = uidRef(vertex.id)
        const hoverId = this.nextObjectId()
        const markupId = this.nextObjectId()
        const normalized = normalizeHover(vertex.result)
        this.rdfStream.write(`${id} <HoverResult.result> ${hoverId} .\n`)
        this.rdfStream.write(`${hoverId} <Hover.contents> ${markupId} .\n`)
        this.rdfStream.write(`${markupId} <MarkupContent.kind> "markdown" .\n`)
        this.rdfStream.write(`${markupId} <MarkupContent.value> ${makeValue(normalized)} .\n`)
    }

    private handleMetaData(vertex: MetaData): void {
        if (!vertex.projectRoot) {
            // TODO - these should be generic validation errors
            throw Object.assign(new Error(`${VertexLabels.metaData} must have a projectRoot field.`), {
                status: 422,
            })
        }

        const rootUri = new URL(vertex.projectRoot)
        // Needed to properly resolve relative URLs
        if (!rootUri.pathname.endsWith('/')) {
            rootUri.pathname += '/'
        }

        if (!semver.satisfies(vertex.version, '^0.4.0')) {
            // TODO - these should be generic validation errors
            throw Object.assign(new Error(`Unsupported LSIF version ${vertex.version}. Supported ^0.4.3`), {
                status: 422,
            })
        }

        const repoUidRef = uidRef('repository')
        const commitUidRef = uidRef('commit')

        this.rootInfo = {
            rootUri,
            repoUidRef,
            commitUidRef,
        }

        this.rdfStream.write(`${repoUidRef} <Repository.name> ${makeValue(this.repository)} .\n`)
        this.rdfStream.write(`${repoUidRef} <Repository.label> ${makeValue('repository')} .\n`)
        this.rdfStream.write(`${repoUidRef} <${EdgeLabels.contains}> ${commitUidRef} .\n`)
        this.rdfStream.write(`${commitUidRef} <Commit.oid> ${makeValue(this.commit)} .\n`)
        this.rdfStream.write(`${commitUidRef} <Commit.label> ${makeValue('commit')} .\n`)
    }

    private handleMoniker(vertex: Moniker): void {
        const id = uidRef(vertex.id)
        this.rdfStream.write(`${id} <Moniker.scheme> ${makeValue(vertex.scheme)} .\n`)
        this.rdfStream.write(`${id} <Moniker.kind> ${makeValue(vertex.kind)} .\n`)
        this.rdfStream.write(`${id} <Moniker.identifier> ${makeValue(vertex.identifier)} .\n`)
    }

    private handlePackageInformation(vertex: PackageInformation): void {
        const id = uidRef(vertex.id)
        this.rdfStream.write(`${id} <PackageInformation.name> ${makeValue(vertex.name)} .\n`)
        this.rdfStream.write(`${id} <PackageInformation.manager> ${makeValue(vertex.manager)} .\n`)
        this.rdfStream.write(`${id} <PackageInformation.version> ${makeValue(vertex.version)} .\n`)
    }

    private handleRange(vertex: Range): void {
        const id = uidRef(vertex.id)
        this.rdfStream.write(`${id} <Range.startLine> ${makeValue(vertex.start.line)} .\n`)
        this.rdfStream.write(`${id} <Range.startCharacter> ${makeValue(vertex.start.character)} .\n`)
        this.rdfStream.write(`${id} <Range.endLine> ${makeValue(vertex.end.line)} .\n`)
        this.rdfStream.write(`${id} <Range.endCharacter> ${makeValue(vertex.end.character)} .\n`)
    }

    private handleReferenceResult(vertex: ReferenceResult): void {
        // FIXME - no properties
    }

    private handleResultSet(vertex: ResultSet): void {
        // FIXME - no properties
    }

    private handleEdge(edge: Edge): void {
        const inVs = Edge.is1N(edge) ? edge.inVs : [edge.inV]
        for (const inV of inVs || []) {
            this.rdfStream.write(`${uidRef(edge.outV)} <${edge.label}> ${uidRef(inV)} .\n`)

            if (edge.label === EdgeLabels.item) {
                this.rdfStream.write(`${uidRef(edge.document)} <${EdgeLabels.contains}> ${uidRef(inV)} .\n`)
            }
        }
    }

    private idCounter = 0

    private nextObjectId(): string {
        return uidRef(`lsifobject${this.idCounter++}`)
    }
}

function relativeUrl(from: URL, to: URL): string {
    return RelateUrl.relate(from.href, to.href, {
        defaultPorts: {}, // Do not remove default ports
        output: RelateUrl.PATH_RELATIVE, // Do not prefer root-relative URLs if shorter
        removeRootTrailingSlash: false, // Do not remove trailing slash
    })
}

function normalizeHover(hover: lsp.Hover): string {
    return (Array.isArray(hover.contents) ? hover.contents : [hover.contents])
        .map(normalizeContent)
        .filter(s => !!s.trim())
        .join('\n\n---\n\n')
}

function normalizeContent(content: string | lsp.MarkupContent | { language: string; value: string }): string {
    if (typeof content === 'string') {
        return content
    }

    if (lsp.MarkupContent.is(content)) {
        return content.value
    }

    if (content.value) {
        return '```' + `${content.language}\n${content.value}\n` + '```'
    }

    return ''
}

function uidRef(id: string | number): string {
    return `_:${id}`
}

function makeValue(value: any): string {
    if (typeof value === 'string') {
        return JSON.stringify(value)
    }

    if (typeof value === 'boolean') {
        return `"${value}^^<xs:boolean>`
    }

    if (typeof value === 'number') {
        return `"${value}"^^<xs:float>`
    }

    return '"<UNKNOWN>"'
}
