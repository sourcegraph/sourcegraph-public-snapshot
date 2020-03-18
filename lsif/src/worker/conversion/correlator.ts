import * as sqliteModels from '../../shared/models/sqlite'
import * as lsif from 'lsif-protocol'
import { createSilentLogger } from '../../shared/logging'
import { DefaultMap } from '../../shared/datastructures/default-map'
import { DisjointSet } from '../../shared/datastructures/disjoint-set'
import { Hover, MarkupContent } from 'vscode-languageserver-types'
import { Logger } from 'winston'
import { mustGet, mustGetFromEither } from '../../shared/maps'
import { relativePath } from './paths'

/** Identifiers of result set vertices. */
export type ResultSetId = lsif.Id

/**
 * An internal representation of a result set vertex. This is only used during
 * correlation and import as we flatten this data into the range vertices for
 * faster queries.
 */
export interface ResultSetData {
    /** The identifier of the definition result attached to this result set. */
    definitionResultId?: sqliteModels.DefinitionResultId

    /** The identifier of the reference result attached to this result set. */
    referenceResultId?: sqliteModels.ReferenceResultId

    /** The identifier of the hover result attached to this result set. */
    hoverResultId?: sqliteModels.HoverResultId

    /** The set of moniker identifiers directly attached to this result set. */
    monikerIds: Set<sqliteModels.MonikerId>
}

/**
 * Common state around the conversion of a single LSIF dump upload. This class
 * receives the parsed vertex or edge, line by line and adds it into an in-memory
 * adjacency-list graph structure that is later processed and converted into a
 * SQLite database on disk.
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
    public documentPaths = new Map<sqliteModels.DocumentId, string>()
    public rangeData = new Map<lsif.RangeId, sqliteModels.RangeData>()
    public resultSetData = new Map<ResultSetId, ResultSetData>()
    public hoverData = new Map<sqliteModels.HoverResultId, string>()
    public monikerData = new Map<sqliteModels.MonikerId, sqliteModels.MonikerData>()
    public packageInformationData = new Map<sqliteModels.PackageInformationId, sqliteModels.PackageInformationData>()
    public unsupportedVertexes = new Set<lsif.Id>()

    // Edge data
    public nextData = new Map<lsif.RangeId | ResultSetId, ResultSetId>()
    public containsData = new Map<sqliteModels.DocumentId, Set<lsif.RangeId>>()
    public definitionData = new Map<
        sqliteModels.DefinitionResultId,
        DefaultMap<sqliteModels.DocumentId, lsif.RangeId[]>
    >()
    public referenceData = new Map<
        sqliteModels.ReferenceResultId,
        DefaultMap<sqliteModels.DocumentId, lsif.RangeId[]>
    >()

    /** A disjoint set of monikers linked by `nextMoniker` edges. */
    public linkedMonikers = new DisjointSet<sqliteModels.MonikerId>()

    /** A disjoint set of reference results linked by `item` edges. */
    public linkedReferenceResults = new DisjointSet<sqliteModels.ReferenceResultId>()

    /** The set of exported moniker identifiers that have package information attached. */
    public importedMonikers = new Set<sqliteModels.MonikerId>()

    /** The set of exported moniker identifiers that have package information attached. */
    public exportedMonikers = new Set<sqliteModels.MonikerId>()

    /**
     * Creates a new Correlator.
     *
     * @param dumpRoot The repository-relative root of all files that are in the dump.
     * @param logger The logger instance.
     */
    constructor(private dumpRoot: string = '', private logger: Logger = createSilentLogger()) {}

    /**
     * Process a single vertex or edge.
     *
     * @param element A vertex or edge element from the LSIF dump.
     */
    public insert(element: lsif.Vertex | lsif.Edge): void {
        if (element.type === lsif.ElementTypes.vertex) {
            switch (element.label) {
                case lsif.VertexLabels.metaData:
                    this.handleMetaData(element)
                    break

                case lsif.VertexLabels.document: {
                    if (!this.projectRoot) {
                        throw new Error('No metadata defined.')
                    }

                    const path = relativePath(this.projectRoot, new URL(element.uri))
                    this.documentPaths.set(element.id, path)
                    this.containsData.set(element.id, new Set<lsif.RangeId>())
                    break
                }

                // The remaining vertex handlers stash data into an appropriate map. This data
                // may be retrieved when an edge that references it is seen, or when a document
                // is finalized.

                case lsif.VertexLabels.range:
                    this.rangeData.set(element.id, {
                        startLine: element.start.line,
                        startCharacter: element.start.character,
                        endLine: element.end.line,
                        endCharacter: element.end.character,
                        monikerIds: new Set<sqliteModels.MonikerId>(),
                    })
                    break

                case lsif.VertexLabels.resultSet:
                    this.resultSetData.set(element.id, { monikerIds: new Set<sqliteModels.MonikerId>() })
                    break

                case lsif.VertexLabels.definitionResult:
                    this.definitionData.set(
                        element.id,
                        new DefaultMap<sqliteModels.DocumentId, lsif.RangeId[]>(() => [])
                    )
                    break

                case lsif.VertexLabels.referenceResult:
                    this.referenceData.set(
                        element.id,
                        new DefaultMap<sqliteModels.DocumentId, lsif.RangeId[]>(() => [])
                    )
                    break

                case lsif.VertexLabels.hoverResult:
                    this.hoverData.set(element.id, normalizeHover(element.result))
                    break

                case lsif.VertexLabels.moniker:
                    this.monikerData.set(element.id, {
                        kind: element.kind || lsif.MonikerKind.local,
                        scheme: element.scheme,
                        identifier: element.identifier,
                    })
                    break

                case lsif.VertexLabels.packageInformation:
                    this.packageInformationData.set(element.id, {
                        name: element.name,
                        version: element.version || null,
                    })
                    break

                default:
                    // Some vertex labels are not yet supported:
                    //
                    // - typeDefinitionResult
                    // - implementationResult
                    // - ... others in the future
                    //
                    // We keep track of these unsupported vertexes so that we
                    // don't mistake it for a missing vertex later when visiting
                    // edges.
                    this.unsupportedVertexes.add(element.id)
                    break
            }
        }

        if (element.type === lsif.ElementTypes.edge) {
            switch (element.label) {
                case lsif.EdgeLabels.contains:
                    this.handleContains(element)
                    break

                case lsif.EdgeLabels.next:
                    this.handleNextEdge(element)
                    break

                case lsif.EdgeLabels.item:
                    this.handleItemEdge(element)
                    break

                case lsif.EdgeLabels.textDocument_definition:
                    this.handleDefinitionEdge(element)
                    break

                case lsif.EdgeLabels.textDocument_references:
                    this.handleReferenceEdge(element)
                    break

                case lsif.EdgeLabels.textDocument_hover:
                    this.handleHoverEdge(element)
                    break

                case lsif.EdgeLabels.moniker:
                    this.handleMonikerEdge(element)
                    break

                case lsif.EdgeLabels.nextMoniker:
                    this.handleNextMonikerEdge(element)
                    break

                case lsif.EdgeLabels.packageInformation:
                    this.handlePackageInformationEdge(element)
                    break
            }
        }
    }

    //
    // Vertex Handlers

    /**
     * This should be the first vertex seen. Extract the project root so we
     * can create relative paths for documents and cache the LSIF protocol
     * version that we will later insert into he metadata table.
     *
     * @param vertex The metadata vertex.
     */
    private handleMetaData({ version, projectRoot }: lsif.MetaData): void {
        this.lsifVersion = version
        this.projectRoot = new URL(projectRoot.endsWith('/') ? projectRoot : projectRoot + '/')

        // We assume that the project root in the LSIF dump is either:
        //
        //   (1) the root of the LSIF dump, or
        //   (2) the root of the repository
        //
        // These are the common cases and we don't explicitly support
        // anything else. Here we normalize to (1) by appending the dump
        // root if it's not already suffixed by it.

        if (this.dumpRoot !== '' && !this.projectRoot.href.endsWith(this.dumpRoot)) {
            this.projectRoot = new URL(this.dumpRoot, `${this.projectRoot.href}/`)
        }
    }

    //
    // Edge Handlers

    /**
     * Add range data ids into the document in which they are contained. Ensures
     * all referenced vertices are defined.
     *
     * @param edge The contains edge.
     */
    private handleContains(edge: lsif.contains): void {
        // Do not track project contains
        if (!this.documentPaths.has(edge.outV)) {
            return
        }

        const set = mustGet(this.containsData, edge.outV, 'contains')
        for (const inV of edge.inVs) {
            mustGet(this.rangeData, inV, 'range')
            set.add(inV)
        }
    }

    /**
     * Update definition and reference fields from an item edge. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The item edge.
     */
    private handleItemEdge(edge: lsif.item): void {
        if (this.definitionData.has(edge.outV)) {
            const documentMap = mustGet(this.definitionData, edge.outV, 'definitionResult')
            const rangeIds = documentMap.getOrDefault(edge.document)
            for (const inV of edge.inVs) {
                mustGet(this.rangeData, inV, 'range')
                rangeIds.push(inV)
            }

            return
        }

        if (this.referenceData.has(edge.outV)) {
            const documentMap = mustGet(this.referenceData, edge.outV, 'referenceResult')
            const rangeIds = documentMap.getOrDefault(edge.document)
            for (const inV of edge.inVs) {
                if (this.referenceData.has(inV)) {
                    this.linkedReferenceResults.union(edge.outV, inV)
                } else {
                    mustGet(this.rangeData, inV, 'range')
                    rangeIds.push(inV)
                }
            }

            return
        }

        if (this.unsupportedVertexes.has(edge.outV)) {
            this.logger.debug('Skipping edge from an unsupported vertex', { edge })
            return
        }

        throw new Error(`Item edge references a nonexistent vertex ${JSON.stringify(edge)}`)
    }

    /**
     * Attaches the specified moniker to the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The moniker edge.
     */
    private handleMonikerEdge(edge: lsif.moniker): void {
        const source = mustGetFromEither<lsif.RangeId, sqliteModels.RangeData, ResultSetId, ResultSetData>(
            this.rangeData,
            this.resultSetData,
            edge.outV,
            'range/resultSet'
        )

        mustGet(this.monikerData, edge.inV, 'moniker')
        source.monikerIds = new Set<sqliteModels.MonikerId>([edge.inV])
    }

    /**
     * Sets the next field of the specified range or result set. Ensures all referenced vertices
     * are defined.
     *
     * @param edge The next edge.
     */
    private handleNextEdge(edge: lsif.next): void {
        mustGetFromEither<lsif.RangeId, sqliteModels.RangeData, ResultSetId, ResultSetData>(
            this.rangeData,
            this.resultSetData,
            edge.outV,
            'range/resultSet'
        )

        mustGet(this.resultSetData, edge.inV, 'resultSet')
        this.nextData.set(edge.outV, edge.inV)
    }

    /**
     * Correlates monikers together so that when one moniker is queried, each correlated moniker
     * is also returned as a strongly connected set. Ensures all referenced vertices are defined.
     *
     * @param edge The nextMoniker edge.
     */
    private handleNextMonikerEdge(edge: lsif.nextMoniker): void {
        mustGet(this.monikerData, edge.inV, 'moniker')
        mustGet(this.monikerData, edge.outV, 'moniker')
        this.linkedMonikers.union(edge.inV, edge.outV)
    }

    /**
     * Sets the package information of the specified moniker. If the moniker is an export moniker,
     * then the package information will also be returned as an exported package by the `finalize`
     * method. Ensures all referenced vertices are defined.
     *
     * @param edge The packageInformation edge.
     */
    private handlePackageInformationEdge(edge: lsif.packageInformation): void {
        const source = mustGet(this.monikerData, edge.outV, 'moniker')
        mustGet(this.packageInformationData, edge.inV, 'packageInformation')
        source.packageInformationId = edge.inV

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
    private handleDefinitionEdge(edge: lsif.textDocument_definition): void {
        const outV = mustGetFromEither<lsif.RangeId, sqliteModels.RangeData, ResultSetId, ResultSetData>(
            this.rangeData,
            this.resultSetData,
            edge.outV,
            'range/resultSet'
        )

        mustGet(this.definitionData, edge.inV, 'definitionResult')
        outV.definitionResultId = edge.inV
    }

    /**
     * Sets the reference result of the specified range or result set. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The textDocument/references edge.
     */
    private handleReferenceEdge(edge: lsif.textDocument_references): void {
        const outV = mustGetFromEither<lsif.RangeId, sqliteModels.RangeData, ResultSetId, ResultSetData>(
            this.rangeData,
            this.resultSetData,
            edge.outV,
            'range/resultSet'
        )

        mustGet(this.referenceData, edge.inV, 'referenceResult')
        outV.referenceResultId = edge.inV
    }

    /**
     * Sets the hover result of the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The textDocument/hover edge.
     */
    private handleHoverEdge(edge: lsif.textDocument_hover): void {
        const outV = mustGetFromEither<lsif.RangeId, sqliteModels.RangeData, ResultSetId, ResultSetData>(
            this.rangeData,
            this.resultSetData,
            edge.outV,
            'range/resultSet'
        )

        mustGet(this.hoverData, edge.inV, 'hoverResult')
        outV.hoverResultId = edge.inV
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
