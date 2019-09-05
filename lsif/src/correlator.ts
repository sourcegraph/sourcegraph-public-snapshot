import RelateUrl from 'relateurl'
import { assertDefined } from './util'
import { DefaultMap } from './default-map'
import { Hover, MarkupContent } from 'vscode-languageserver-types'
import { MonikerData, PackageInformationData, RangeData } from './entities'
import {
    Id,
    VertexLabels,
    EdgeLabels,
    Vertex,
    Edge,
    MonikerKind,
    ItemEdgeProperties,
    moniker,
    next,
    nextMoniker,
    textDocument_definition,
    textDocument_hover,
    textDocument_references,
    packageInformation,
    item,
    MetaData,
    ElementTypes,
    contains,
} from 'lsif-protocol'

/**
 * An internal representation of a result set vertex. This is only used during import
 * as we flatten this data into the range vertices for faster queries.
 */
export interface ResultSetData {
    /**
     * * The identifier of the definition result attached to this result set.
     */
    definitionResult?: Id

    /**
     * * The identifier of the reference result attached to this result set.
     */
    referenceResult?: Id

    /**
     * * The identifier of the hover result attached to this result set.
     */
    hoverResult?: Id

    /**
     * * The set of moniker identifiers directly attached to this result set.
     */
    monikers: Id[]
}

/**
 * Common state around the conversion of a single LSIF dump upload. This class
 * receives the parsed vertex or edge, line by line, from the caller, and adds it
 * into an in-memory structure that is later processed and converted into a SQLite
 * database on disk.
 */
export class Correlator {
    /**
     * The LSIF version of the input. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    public lsifVersion?: string

    /**
     * The root of all document URIs. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    public projectRoot?: URL

    // Vertex data
    public documentPaths = new Map<Id, string>()
    public rangeData = new Map<Id, RangeData>()
    public resultSetData = new Map<Id, ResultSetData>()
    public hoverData = new Map<Id, string>()
    public monikerData = new Map<Id, MonikerData>()
    public packageInformationData = new Map<Id, PackageInformationData>()

    // Edge data
    public nextData = new Map<Id, Id>()
    public containsData = new Map<Id, Set<Id>>() // document to ranges
    public definitionData = new Map<Id, DefaultMap<Id, Id[]>>() // definition result to document to ranges
    public referenceData = new Map<Id, DefaultMap<Id, Id[]>>() // reference result to document to ranges

    /**
     * A mapping for the relation from moniker to the set of monikers that they are related
     * to via nextMoniker edges. This relation is symmetric (if `a` is in `MonikerSets[b]`,
     * then `b` is in `monikerSets[a]`).
     */
    public monikerSets = new DefaultMap<Id, Set<Id>>(() => new Set<Id>())

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    public importedMonikers = new Set<Id>()

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    public exportedMonikers = new Set<Id>()

    /**
     * Process a single vertex or edge.
     *
     * @param element A vertex or edge element from the LSIF dump.
     */
    public insert(element: Vertex | Edge): void {
        if (element.type === ElementTypes.vertex) {
            switch (element.label) {
                case VertexLabels.metaData:
                    this.handleMetaData(element)
                    break

                case VertexLabels.document: {
                    if (!this.projectRoot) {
                        throw new Error('No metadata defined.')
                    }

                    const path = RelateUrl.relate(this.projectRoot.href + '/', new URL(element.uri).href, {
                        defaultPorts: {},
                        output: RelateUrl.PATH_RELATIVE,
                        removeRootTrailingSlash: false,
                    })

                    this.documentPaths.set(element.id, path)
                    this.containsData.set(element.id, new Set<Id>())
                    break
                }

                // The remaining vertex handlers stash data into an appropriate map. This data
                // may be retrieved when an edge that references it is seen, or when a document
                // is finalized.

                case VertexLabels.range:
                    this.rangeData.set(element.id, {
                        startLine: element.start.line,
                        startCharacter: element.start.character,
                        endLine: element.end.line,
                        endCharacter: element.end.character,
                        monikers: [],
                    })
                    break

                case VertexLabels.resultSet:
                    this.resultSetData.set(element.id, { monikers: [] })
                    break

                case VertexLabels.definitionResult:
                    this.definitionData.set(element.id, new DefaultMap<Id, Id[]>(() => []))
                    break

                case VertexLabels.referenceResult:
                    this.referenceData.set(element.id, new DefaultMap<Id, Id[]>(() => []))
                    break

                case VertexLabels.hoverResult:
                    this.hoverData.set(element.id, normalizeHover(element.result))
                    break

                case VertexLabels.moniker:
                    this.monikerData.set(element.id, {
                        kind: element.kind || MonikerKind.local,
                        scheme: element.scheme,
                        identifier: element.identifier,
                    })
                    break

                case VertexLabels.packageInformation:
                    this.packageInformationData.set(element.id, {
                        name: element.name,
                        version: element.version || '$missing',
                    })
                    break
            }
        }

        if (element.type === ElementTypes.edge) {
            switch (element.label) {
                case EdgeLabels.contains:
                    this.handleContains(element)
                    break

                case EdgeLabels.next:
                    this.handleNextEdge(element)
                    break

                case EdgeLabels.item:
                    this.handleItemEdge(element)
                    break

                case EdgeLabels.textDocument_definition:
                    this.handleDefinitionEdge(element)
                    break

                case EdgeLabels.textDocument_references:
                    this.handleReferenceEdge(element)
                    break

                case EdgeLabels.textDocument_hover:
                    this.handleHoverEdge(element)
                    break

                case EdgeLabels.moniker:
                    this.handleMonikerEdge(element)
                    break

                case EdgeLabels.nextMoniker:
                    this.handleNextMonikerEdge(element)
                    break

                case EdgeLabels.packageInformation:
                    this.handlePackageInformationEdge(element)
                    break
            }
        }
    }

    //
    // Vertex Handlers

    /**
     * This should be the first vertex seen. Extract the project root so we
     * can create relative paths for documents. Insert a row in the meta
     * table with the LSIF protocol version.
     *
     * @param vertex The metadata vertex.
     */
    private handleMetaData(vertex: MetaData): void {
        this.lsifVersion = vertex.version
        this.projectRoot = new URL(vertex.projectRoot)
    }

    //
    // Edge Handlers

    /**
     * Add range data ids into the document in which they are contained. Ensures
     * all referenced vertices are defined.
     *
     * @param edge The contains edge.
     */
    private handleContains(edge: contains): void {
        // Do not track project contains
        if (!this.documentPaths.has(edge.outV)) {
            return
        }

        const set = assertDefined(edge.outV, 'contains', this.containsData)
        for (const inV of edge.inVs) {
            assertDefined(inV, 'range', this.rangeData)
            set.add(inV)
        }
    }

    /**
     * Update definition and reference fields from an item edge. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The item edge.
     */
    private handleItemEdge(edge: item): void {
        switch (edge.property) {
            // `item` edges with a `property` refer to a referenceResult
            case ItemEdgeProperties.definitions:
            case ItemEdgeProperties.references: {
                const documentMap = assertDefined(edge.outV, 'referenceResult', this.referenceData)
                const rangeIds = documentMap.getOrDefault(edge.document)
                for (const inV of edge.inVs) {
                    assertDefined(inV, 'range', this.rangeData)
                    rangeIds.push(inV)
                }

                break
            }

            // `item` edges without a `property` refer to a definitionResult
            case undefined: {
                const documentMap = assertDefined(edge.outV, 'definitionResult', this.definitionData)
                const rangeIds = documentMap.getOrDefault(edge.document)
                for (const inV of edge.inVs) {
                    assertDefined(inV, 'range', this.rangeData)
                    rangeIds.push(inV)
                }

                break
            }
        }
    }

    /**
     * Attaches the specified moniker to the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The moniker edge.
     */
    private handleMonikerEdge(edge: moniker): void {
        const source = assertDefined<Id, RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )

        assertDefined(edge.inV, 'moniker', this.monikerData)
        source.monikers = [edge.inV]
    }

    /**
     * Sets the next field of the specified range or result set. Ensures all referenced vertices
     * are defined.
     *
     * @param edge The next edge.
     */
    private handleNextEdge(edge: next): void {
        assertDefined<Id, RangeData | ResultSetData>(edge.outV, 'range/resultSet', this.rangeData, this.resultSetData)
        assertDefined(edge.inV, 'resultSet', this.resultSetData)
        this.nextData.set(edge.outV, edge.inV)
    }

    /**
     * Correlates monikers together so that when one moniker is queried, each correlated moniker
     * is also returned as a strongly connected set. Ensures all referenced vertices are defined.
     *
     * @param edge The nextMoniker edge.
     */
    private handleNextMonikerEdge(edge: nextMoniker): void {
        assertDefined(edge.inV, 'moniker', this.monikerData)
        assertDefined(edge.outV, 'moniker', this.monikerData)
        this.monikerSets.getOrDefault(edge.inV).add(edge.outV) // Forward direction
        this.monikerSets.getOrDefault(edge.outV).add(edge.inV) // Backwards direction
    }

    /**
     * Sets the package information of the specified moniker. If the moniker is an export moniker,
     * then the package information will also be returned as an exported package by the `finalize`
     * method. Ensures all referenced vertices are defined.
     *
     * @param edge The packageInformation edge.
     */
    private handlePackageInformationEdge(edge: packageInformation): void {
        const source = assertDefined(edge.outV, 'moniker', this.monikerData)
        assertDefined(edge.inV, 'packageInformation', this.packageInformationData)
        source.packageInformation = edge.inV

        if (source.kind === 'export') {
            this.exportedMonikers.add(edge.outV)
        }

        if (source.kind === 'import') {
            this.importedMonikers.add(edge.outV)
        }
    }

    /**
     * Sets the definition result of the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The textDocument/definition edge.
     */
    private handleDefinitionEdge(edge: textDocument_definition): void {
        const outV = assertDefined<Id, RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )

        assertDefined(edge.inV, 'definitionResult', this.definitionData)
        outV.definitionResult = edge.inV
    }

    /**
     * Sets the hover result of the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The textDocument/hover edge.
     */
    private handleHoverEdge(edge: textDocument_hover): void {
        const outV = assertDefined<Id, RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )

        assertDefined(edge.inV, 'hoverResult', this.hoverData)
        outV.hoverResult = edge.inV
    }

    /**
     * Sets the reference result of the specified range or result set. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The textDocument/references edge.
     */
    private handleReferenceEdge(edge: textDocument_references): void {
        const outV = assertDefined<Id, RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )

        assertDefined(edge.inV, 'referenceResult', this.referenceData)
        outV.referenceResult = edge.inV
    }
}

/**
 * Normalize an LSP hover object into a string.
 *
 * @param hover The hover object.
 */
export function normalizeHover(hover: Hover): string {
    const normalizeContent = (content: string | MarkupContent | { language: string; value: string }): string => {
        if (typeof content === 'string') {
            return content
        }

        if (MarkupContent.is(content)) {
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
