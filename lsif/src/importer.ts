import * as lsp from 'vscode-languageserver'
import RelateUrl from 'relateurl'
import {
    Document,
    Edge,
    ElementTypes,
    HoverResult,
    Range,
    MetaData,
    Vertex,
    VertexLabels,
    Moniker,
    PackageInformation,
    Id,
    EdgeLabels,
    ItemEdge,
    V,
} from 'lsif-protocol'
import { fs, child_process, readline } from 'mz'
import { Readable } from 'stream'
import { readEnv } from './util'

/**
 * The address used to connect to the Dgraph server for live loading.
 */
const DGRAPH_LIVE_LOADER_ADDRESS = readEnv('DGRAPH_LIVE_LOADER_ADDRESS', 'localhost:5080')

/**
 * `HandlerMap` is a mapping from vertex or edge labels to the function that
 * can handle an object of that particular type during import.
 */
interface HandlerMap {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [K: string]: (element: any) => void
}

/**
 * `Importer` converts a stream of LSIF data into a format readable by Dgraph,
 * and inserts it to the instance running at the configured address.
 */
export class Importer {
    // Handler vtables
    private vertexHandlerMap: HandlerMap = {}
    private edgeHandlerMap: HandlerMap = {}

    /**
     * `projectRoot` is the prefix of all document URIs. This is extracted from
     * the metadata vertex at the beginning of processing.
     */
    private projectRoot: URL | undefined

    /**
     * `stream` is an `fs.WriteStream` to the temporary RDF file generated in the
     * first step of the import.
     */
    private stream!: fs.WriteStream

    /**
     * Create a new `Importer` for the given repository and commit.
     *
     * @param repository The repository.
     * @param commit The commit hash.
     */
    constructor(private repository: string, private commit: string) {
        // Register vertex handlers
        this.vertexHandlerMap[VertexLabels.document] = e => this.handleDocument(e)
        this.vertexHandlerMap[VertexLabels.hoverResult] = e => this.handleHoverResult(e)
        this.vertexHandlerMap[VertexLabels.metaData] = e => this.handleMetaData(e)
        this.vertexHandlerMap[VertexLabels.moniker] = e => this.handleMoniker(e)
        this.vertexHandlerMap[VertexLabels.packageInformation] = e => this.handlePackageInformation(e)
        this.vertexHandlerMap[VertexLabels.range] = e => this.handleRange(e)

        // Register edge handlers
        this.edgeHandlerMap[EdgeLabels.contains] = e => this.handleMultiEdge(e)
        this.edgeHandlerMap[EdgeLabels.item] = e => this.handleItemEdge(e)
        this.edgeHandlerMap[EdgeLabels.moniker] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.next] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.nextMoniker] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.packageInformation] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.textDocument_definition] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.textDocument_hover] = e => this.handleSingleEdge(e)
        this.edgeHandlerMap[EdgeLabels.textDocument_references] = e => this.handleSingleEdge(e)
    }

    /**
     * Import the provided LSIF dump data into Dgraph. This is done in two steps:
     *
     *   1. Convert each JSON line into one or more RDF lines in a temp file.
     *   2. Run the dgraph live loader over the generated RDF input.
     *
     * It is assumed that the provided temporary file path refers to a valid and
     * empty file on disk, and will be cleaned up by the caller.
     *
     * @param input The LSIF dump input stream.
     * @param rdfPath A path to a temporary file to store RDF data.
     */
    public async import(input: Readable, rdfPath: string): Promise<void> {
        this.stream = fs.createWriteStream(rdfPath)

        console.time('converting')
        for await (const line of readline.createInterface({ input })) {
            let element: Vertex | Edge
            try {
                element = JSON.parse(line)
            } catch (e) {
                throw new Error(`Parsing failed for line:\n${line}`)
            }

            try {
                this.handleElement(element)
            } catch (e) {
                throw new Error(`Failed to process line:\n${line}\nCaused by:\n${e}`)
            }
        }
        console.timeEnd('converting')

        console.time('flushing')
        this.stream.end()
        await new Promise(resolve => this.stream.on('finish', resolve))
        console.timeEnd('flushing')

        console.time('loading')

        const proc = child_process.spawn('dgraph', [
            'live',
            '-z',
            DGRAPH_LIVE_LOADER_ADDRESS,
            '-r',
            rdfPath,
            '-C',
            '-c',
            '1',
        ])

        if (proc.stdout && proc.stderr) {
            proc.stdout.on('data', data => console.log('stdout: ' + data.toString()))
            proc.stderr.on('data', data => console.log('stderr: ' + data.toString()))
        }

        await new Promise(resolve => proc.on('exit', resolve))
        console.timeEnd('loading')
    }

    /**
     * Process a single vertex or edge.
     *
     * @param element A vertex or edge element from the LSIF dump.
     */
    private handleElement(element: Vertex | Edge): void {
        const handler =
            element.type === ElementTypes.vertex
                ? this.vertexHandlerMap[element.label]
                : this.edgeHandlerMap[element.label]

        if (handler) {
            handler(element)
        }
    }

    /**
     * Emit a document vertex. This adds a property `path` that is the `uri` relative
     * to the project root. It is a fatal error if the metadata vertex has yet to be
     * visited. This also adds an additional edge to the commit vertex emitted by the
     * handling fo the metadata vertex.
     *
     * @param vertex The vertex object.
     */
    private handleDocument(vertex: Document): void {
        if (!this.projectRoot) {
            throw new Error('No root URI')
        }

        const path = decodeURIComponent(relativeUrl(this.projectRoot, new URL(vertex.uri)))
        this.emitVertex(vertex, 'document', { path: makeValue(path) })
        this.emitEdge('commit', 'contains', vertex.id)
    }

    /**
     * Emit a hover vertex. This collapses the subgraph around hovers into a single
     * layer so we do not need to store as many useless edges.
     *
     * LSIF: HoverResult -> Hover -> MarkupContent{kind, value}
     * Dgraph: hoverResult{kind, value}
     *
     * @param vertex The vertex object.
     */
    private handleHoverResult(vertex: HoverResult): void {
        this.emitVertex(vertex, 'hoverResult', {
            kind: makeValue('markdown'),
            value: makeValue(normalizeHover(vertex.result)),
        })
    }

    /**
     * Emit a metadata vertex. This will set the `rootUri` value for later use when
     * handling document vertices.
     *
     * @param vertex The vertex object.
     */
    private handleMetaData(vertex: MetaData): void {
        this.projectRoot = new URL(vertex.projectRoot)
        if (!this.projectRoot.pathname.endsWith('/')) {
            this.projectRoot.pathname += '/'
        }

        this.emitVertex({ id: 'repository' }, 'repository', { name: makeValue(this.repository) })
        this.emitVertex({ id: 'commit' }, 'commit', { revhash: makeValue(this.commit) })
        this.emitEdge('repository', 'contains', 'commit')
    }

    /**
     * Emit a moniker vertex.
     *
     * @param vertex The vertex object.
     */
    private handleMoniker(vertex: Moniker): void {
        this.emitVertex(vertex, 'moniker', {
            scheme: makeValue(vertex.scheme),
            kind: makeValue(vertex.kind),
            identifier: makeValue(vertex.identifier),
        })
    }

    /**
     * Emit a packageInformation vertex.
     *
     * @param vertex The vertex object.
     */
    private handlePackageInformation(vertex: PackageInformation): void {
        this.emitVertex(vertex, 'packageInformation', {
            name: makeValue(vertex.name),
            manager: makeValue(vertex.manager),
            version: makeValue(vertex.version),
        })
    }

    /**
     * Emit a range vertex. This flattens the start and end values into the top-level
     * of the vertex so it can be queried efficiently.
     *
     * @param vertex The vertex object.
     */
    private handleRange(vertex: Range): void {
        this.emitVertex(vertex, 'range', {
            startLine: makeValue(vertex.start.line),
            startCharacter: makeValue(vertex.start.character),
            endLine: makeValue(vertex.end.line),
            endCharacter: makeValue(vertex.end.character),
        })
    }

    /**
     * Emit a 1:1 edge.
     *
     * @param edge The edge object.
     */
    private handleSingleEdge(edge: { label: string; outV: string; inV: string }): void {
        this.emitEdge(edge.outV, edge.label, edge.inV)
    }

    /**
     * Emit a 1:n edge.
     *
     * @param edge The edge object.
     */
    private handleMultiEdge(edge: { label: string; outV: string; inVs: string[] }): void {
        for (const inV of edge.inVs) {
            this.emitEdge(edge.outV, edge.label, inV)
        }
    }

    /**
     * Emit an item edge. This additionally has a `document` property that can help fill
     * out the contains relation.
     *
     * @param edge The edge object.
     */
    private handleItemEdge<S extends V, T extends V>(edge: ItemEdge<S, T>): void {
        for (const inV of edge.inVs) {
            this.emitEdge(edge.outV, edge.label, inV)
            this.emitEdge(edge.document, 'contains', inV)
        }
    }

    /**
     * Emit the properties of a vertex. Vertex properties are also stored as a predicate
     * in Dgraph, so we generate each one separately here. Each vertex also has a special
     * `label` property that allows queries to determine the type of vertex during graph
     * traversal.
     *
     * @param vertex The vertex object.
     * @param name The type of the vertex.
     * @param properties A map of vertex properties to emit.
     */
    private emitVertex(vertex: { id: Id }, name: string, properties: { [K: string]: string }): void {
        const id = `_:${vertex.id}`
        this.emitLine(id, `${name}.label`, makeValue(name))

        for (const property of Object.keys(properties)) {
            this.emitLine(id, `${name}.${property}`, properties[property])
        }
    }

    /**
     * Emit an edge relation between two vertices.
     *
     * @param outV The source vertex identifier.
     * @param predicate The relation between vertices.
     * @param inV The destination vertex identifier.
     */
    private emitEdge(outV: Id, predicate: string, inV: Id): void {
        this.emitLine(`_:${outV}`, predicate, `_:${inV}`)
    }

    /**
     * Emit a raw line of the form `<subject> <predicate> <object> .`.
     *
     * @param subject The subject - must be a vertex id.
     * @param predicate The predicate.
     * @param object The object - can be a vertex id or a value created with `makeValue`.
     */
    private emitLine(subject: string, predicate: string, object: string): void {
        this.stream.write(`${subject} <${predicate}> ${object} .\n`)
    }
}

function relativeUrl(from: URL, to: URL): string {
    return RelateUrl.relate(from.href, to.href, {
        defaultPorts: {}, // Do not remove default ports
        output: RelateUrl.PATH_RELATIVE, // Do not prefer root-relative URLs if shorter
        removeRootTrailingSlash: false, // Do not remove trailing slash
    })
}

/**
 * Normalize an LSP hover object into a string.
 *
 * @param hover The hover object.
 */
function normalizeHover(hover: lsp.Hover): string {
    const normalizeContent = (content: string | lsp.MarkupContent | { language: string; value: string }): string => {
        if (typeof content === 'string') {
            return content
        }

        if (lsp.MarkupContent.is(content)) {
            return content.value
        }

        const tick = '```'
        return `${tick}${content.language}\n${content.value}\n${tick}`
    }

    const separator = '\n\n---\n\n'
    const contents = Array.isArray(hover.contents) ? hover.contents : [hover.contents]
    return contents
        .map(c => normalizeContent(c).trim())
        .filter(s => s)
        .join(separator)
}

/**
 * Convert a TypeScript value into a Dgraph literal.
 *
 * @param value The value to convert.
 */
function makeValue(value: string | boolean | number | undefined): string {
    if (typeof value === 'boolean') {
        return `"${value}"^^<xs:boolean>`
    }

    if (typeof value === 'number') {
        return `"${Math.floor(value)}"^^<xs:int>`
    }

    return JSON.stringify(value)
}
