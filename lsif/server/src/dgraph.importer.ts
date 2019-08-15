import * as lsp from 'vscode-languageserver'
import * as semver from 'semver'
import RelateUrl from 'relateurl'
import { Document, Edge, EdgeLabels, ElementTypes, HoverResult, Range, Vertex, VertexLabels } from 'lsif-protocol'
import { Hover, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { MetaData } from './ms/server/protocol.compress'
import { Mutation, NQuad, Value, DgraphClient } from 'dgraph-js'
import { Schema } from 'jsonschema'
import { validate } from './dgraph.schema'
import { flattenRange } from './dgraph.range'

/**
 * Inserts LSIF dump data into Dgraph.
 */
export class DgraphImporter {
    private mutation: Mutation
    private idCounter = 0
    private rootInfo: { rootUri: URL; repoUidRef: string; commitUidRef: string } | undefined

    constructor(
        private client: DgraphClient,
        private repository: string,
        private commit: string,
        private schema: Schema
    ) {
        this.mutation = new Mutation()
    }

    /**
     * Create a mutation for each vertex and edge in the given list and apply
     * it to a transaction created in the constructor. This method will ensure
     * cleanup of the transaction.
     */
    public async import(items: (Vertex | Edge)[]): Promise<void> {
        console.log('doing mutation setup')
        console.time('setup')

        items.forEach(element => {
            if (element.type === ElementTypes.vertex) {
                this.insertVertex(element)
            } else {
                this.insertEdge(element)
            }
        })

        console.timeEnd('setup')

        console.log('nquads:', this.mutation.getSetList().length)
        console.log('calling mutate')

        const transaction = this.client.newTxn()
        try {
            console.time('mutate')
            const assigned = await transaction.mutate(this.mutation)
            console.timeEnd('mutate')

            console.log('calling commit')
            console.time('commit')
            await transaction.commit()
            console.timeEnd('commit')
            console.log(assigned.toObject())
        } finally {
            await transaction.discard()
        }
    }

    /**
     * Insert data related to the given edge into the mutation.
     */
    private insertEdge(element: Edge): void {
        // 1:n edges get flattened into multiple 1:1 edges
        const inVs = Edge.is1N(element) ? element.inVs : [element.inV]
        for (const inV of inVs || []) {
            const nquad = makeTriplet(uidRef(element.outV), element.label, uidRef(inV))
            this.mutation.addSet(nquad)
            if (element.label === EdgeLabels.item) {
                // Make sure the ranges linked by item are linked to the document with a contains
                // Otherwise it would not be possible to determine which document result ranges belong too
                // This depends on contains having reverse edges enabled
                this.mutation.addSet(makeTriplet(uidRef(element.document), EdgeLabels.contains, uidRef(inV)))
            }
        }
    }

    /**
     * Insert data related to te given vertex into the mutation.
     */
    private insertVertex(element: Vertex): void {
        if (element.label === VertexLabels.event) {
            return
        }

        if (element.label === VertexLabels.metaData) {
            this.insertMetadata(element)
            return
        }

        if (!this.rootInfo) {
            // TODO - these should be generic validation errors
            throw Object.assign(
                new Error('The first vertex must be a metaData vertex that specifies the projectRoot URI'),
                { status: 422 }
            )
        }

        const { type, ...rest } = element
        const vertexId = String(element.id)
        const vertexProperties = this.getVertexProperties(element, rest)

        if (element.label === VertexLabels.document) {
            // Add contains edge from commit to document
            this.mutation.addSet(makeTriplet(this.rootInfo.commitUidRef, EdgeLabels.contains, uidRef(vertexId)))
        }

        // Add subgraph for vertex properties
        this.addObject(
            null,
            vertexProperties,
            vertexId,
            validate(this.schema.definitions!.Vertex, {
                type,
                ...vertexProperties,
            })
        )
    }

    /**
     * Insert data associated with the project. This is assumed to be at the top
     * level of the dump. This method populates the `rootInfo` of the object, which
     * may be used by other methods to construct relative URIs.
     */
    private insertMetadata(element: MetaData): void {
        if (!element.projectRoot) {
            // TODO - these should be generic validation errors
            throw Object.assign(new Error(`${VertexLabels.metaData} must have a projectRoot field.`), {
                status: 422,
            })
        }

        const rootUri = new URL(element.projectRoot)
        // Needed to properly resolve relative URLs
        if (!rootUri.pathname.endsWith('/')) {
            rootUri.pathname += '/'
        }

        if (!semver.satisfies(element.version, '^0.4.0')) {
            // TODO - these should be generic validation errors
            throw Object.assign(new Error(`Unsupported LSIF version ${element.version}. Supported ^0.4.3`), {
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

        // Create repository and commit vertexes (not part of LSIF spec)
        this.mutation.addSet(makeTriplet(repoUidRef, 'Repository.name', makeValue(this.repository)))
        this.mutation.addSet(makeTriplet(repoUidRef, 'Repository.label', makeValue('repository')))
        this.mutation.addSet(makeTriplet(repoUidRef, EdgeLabels.contains, commitUidRef))
        this.mutation.addSet(makeTriplet(commitUidRef, 'Commit.oid', makeValue(this.commit)))
        this.mutation.addSet(makeTriplet(commitUidRef, 'Commit.label', makeValue('commit')))
    }

    /**
     * Explicitly construct metadata for a generic vertex.
     */
    private getVertexProperties(element: Vertex, defaultProperties: object): object {
        if (element.label === VertexLabels.document) {
            return this.getDocumentVertexProperties(element)
        } else if (element.label === VertexLabels.hoverResult) {
            return this.getHoverResultVertexProperties(element)
        } else if (element.label === VertexLabels.range) {
            return this.getRangeVertexProperties(element)
        }

        return defaultProperties
    }

    /**
     * Explicitly construct metadata for a document vertex.
     */
    private getDocumentVertexProperties(element: Document): object {
        // Ignore content for now to not bloat DB
        const { uri, contents, ...rest } = element
        // Store root-relative path instead of URI to exclude context from the local machine
        const path = decodeURIComponent(relativeUrl(this.rootInfo!.rootUri, new URL(uri)))
        return { ...rest, path }
    }

    /**
     * Explicitly construct metadata for a hover result vertex.
     */
    private getHoverResultVertexProperties(element: HoverResult): object {
        // normalize to MarkupContent to remove union types
        return { ...element, result: normalizeHover(element.result) }
    }

    /**
     * Explicitly construct metadata for a range vertex.
     */
    private getRangeVertexProperties(element: Range): object {
        // Flatten range properties so it's possible to query them
        const { start, end, ...rest } = element
        return { ...rest, ...flattenRange(element) }
    }

    /**
     * TODO - document
     */
    private addObject(parentConnection: ParentConnection | null, obj: object, objId: string, schema: Schema): void {
        // if (!schema.title) {
        //     throw Object.assign(new Error('Passed schema has no title'), { schema, obj })
        // }

        // For objects, start a new node and add NQuads for its properties recursively
        // If there is a parent, link from it
        if (parentConnection) {
            this.mutation.addSet(
                makeTriplet(uidRef(parentConnection.subjectId), parentConnection.predicate, uidRef(objId))
            )
        }

        for (const [propertyName, propertyValue] of Object.entries(obj)) {
            // TODO - this is the only place that any schema is used
            // Can we get a substitutable value that doesn't require
            // any of the schema crud?

            this.addAny(
                { subjectId: objId, predicate: `${schema.title}.${propertyName}` },
                propertyValue,
                schema.properties![propertyName]
            )
        }
    }

    /**
     * TODO - document
     */
    private addAny(parentConnection: ParentConnection, value: unknown, schema: Schema): void {
        // Ignore undefined or null properties
        if (value === undefined || value === null) {
            return
        }

        if (typeof value === 'object' && value !== null) {
            // For arrays, create multiple edges from object to every array item
            // CAVEAT 1: The order of array items get lost here.
            // In general, LSP arrays are unordered sets.
            // If a list is not, we could add the index as a facet.
            // CAVEAT 2: Nested lists get flattened.
            // I am not aware of any nested lists in LSIF.
            // Possible alternative: create a node for the array.
            if (Array.isArray(value)) {
                for (const [index, item] of value.entries()) {
                    const itemSchema = (schema.items as Schema[])[index]
                    this.addAny(parentConnection, item, itemSchema)
                }

                return
            }

            this.addObject(parentConnection, value, this.nextObjectId(), schema)
            return
        }

        // For scalars, create an edge from object to scalar value
        this.mutation.addSet(
            makeTriplet(uidRef(parentConnection.subjectId), parentConnection.predicate, makeValue(value))
        )
    }

    /**
     * Generate a unique string identifier.
     */
    private nextObjectId(): string {
        return `lsifobject${this.idCounter++}`
    }
}

/**
 * An object used to track contains relationships as mutations as
 * the predicates of an LSIF dump are read.
 */
interface ParentConnection {
    subjectId: string
    predicate: string
}

/**
 * Returns the relative URL from `from` to `to` (the inverse of `url.resolve`).
 */
function relativeUrl(from: URL, to: URL): string {
    return RelateUrl.relate(from.href, to.href, {
        defaultPorts: {}, // Do not remove default ports
        output: RelateUrl.PATH_RELATIVE, // Do not prefer root-relative URLs if shorter
        removeRootTrailingSlash: false, // Do not remove trailing slash
    })
}

/**
 * Ensure that the LSP hover object always uses MarkupContent as Dgraph does
 * not have support for union types.
 */
function normalizeHover(hover: Hover): Hover {
    const normalized = (Array.isArray(hover.contents) ? hover.contents : [hover.contents])
        .map(normalizeContent)
        .filter(s => !!s.trim())
        .join('\n\n---\n\n')

    return { ...hover, contents: { kind: MarkupKind.Markdown, value: normalized } }
}

/**
 * Convert (possibly) marked content into a bare string.
 */
function normalizeContent(content: string | lsp.MarkupContent | { language: string; value: string }): string {
    if (typeof content === 'string') {
        return content
    }

    if (MarkupContent.is(content)) {
        return content.value
    }

    if (content.value) {
        return '```' + `${content.language}\n${content.value}\n` + '```'
    }

    return ''
}

/**
 * Create a placeholder UID reference.
 */
function uidRef(id: string | number): string {
    return '_:' + id
}

/**
 * Create a Dgraph triplet from the subject UID, a globally unique predicate
 * (this is `Type.property` by convention), and a UID or scalar object value.
 */
function makeTriplet(subject: string, predicate: string, object: Value | string): NQuad {
    const nquad = new NQuad()
    nquad.setSubject(subject)
    nquad.setPredicate(predicate)

    if (typeof object === 'string') {
        nquad.setObjectId(object)
    } else {
        nquad.setObjectValue(object)
    }

    return nquad
}

/**
 * Convert a TS scalar value into a Dgraph value.
 */
function makeValue(value: any): Value {
    const graphValue = new Value()

    if (typeof value === 'number') {
        if (Number.isInteger(value)) {
            graphValue.setIntVal(value)
        } else {
            graphValue.setDoubleVal(value)
        }
    } else if (typeof value === 'boolean') {
        graphValue.setBoolVal(value)
    } else if (typeof value === 'string') {
        graphValue.setStrVal(value)
    }

    return graphValue
}
