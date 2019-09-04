import { assertDefined, assertId, hashKey } from './util'
import { Correlator, ResultSetData } from './correlator'
import { DefaultMap } from './default-map'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ChunkModel } from './models.database'
import { DocumentData, MonikerData, PackageInformationData, RangeData, QualifiedRangeId } from './entities'
import { Edge, Id, MonikerKind, Vertex } from 'lsif-protocol'
import { encodeJSON } from './encoding'
import { EntityManager } from 'typeorm'
import { isEqual, uniqWith } from 'lodash'
import { Package, SymbolReferences } from './xrepo'
import { TableInserter } from './inserter'

/**
 * The internal version of our SQLite databases. We need to keep this in case
 * we add something that can't be done transparently; if we change how we process
 * something in the future we'll need to consider a number of previous version
 * while we update or re-process the already-uploaded data.
 */
const INTERNAL_LSIF_VERSION = '0.1.0'

/**
 * Correlate each vertex and edge together, then populate the provided entity manager
 * with the document, definition, and reference information. Returns the package and
 * external reference data needed to populate the correlation database.
 *
 * @param entityManager A transactional SQLite entity manager.
 * @param elements The stream of vertex and edge objects composing the LSIF dump.
 */
export async function importLsif(
    entityManager: EntityManager,
    elements: AsyncIterable<Vertex | Edge>
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    const correlator = new Correlator()

    console.time('the whole shebang')

    console.time('inserting')
    let line = 0
    for await (const element of elements) {
        try {
            correlator.insert(element)
        } catch (e) {
            // TODO - more context
            throw Object.assign(new Error(`Failed to process line:\n${line}`), { e })
        }

        line++
    }
    console.timeEnd('inserting')

    if (correlator.lsifVersion === undefined) {
        throw new Error('No metadata defined.')
    }

    // Determine the max batch size of each model type. We cannot perform an
    // insert operation with more than 999 placeholder variables, so we need
    // to flush our batch before we reach that amount. The batch size for each
    // model is calculated based on the number of fields inserted. If fields
    // are added to the models, these numbers will also need to change.

    const metaInserter = new TableInserter(entityManager, MetaModel, Math.floor(999 / 3))
    const documentInserter = new TableInserter(entityManager, DocumentModel, Math.floor(999 / 2))
    const chunkInserter = new TableInserter(entityManager, ChunkModel, Math.floor(999 / 2))
    const definitionInserter = new TableInserter(entityManager, DefinitionModel, Math.floor(999 / 8))
    const referenceInserter = new TableInserter(entityManager, ReferenceModel, Math.floor(999 / 8))

    // Insert uploaded LSIF and the current version of the importer
    await metaInserter.insert({
        lsifVersion: correlator.lsifVersion,
        sourcegraphVersion: INTERNAL_LSIF_VERSION,
    })

    console.time('canonicalizing items')
    for (const [id, range] of correlator.rangeData) {
        canonicalizeItem(correlator, id, range)
    }
    console.timeEnd('canonicalizing items')

    console.time('gathering documents')
    for (const [documentId, documentPath] of correlator.documentPaths) {
        // Create document record from the correlated information. This will also insert
        // external definitions and references into the maps initialized above, which are
        // inserted into the definitions and references table, respectively, below.

        const document = gatherDocument(
            correlator,
            documentId,
            documentPath
            // definitions.getOrDefault(documentId),
            // references.getOrDefault(documentId)
        )

        // Encode and insert insert document
        await documentInserter.insert({
            path: documentPath,
            data: await encodeJSON({
                ranges: document.ranges,
                orderedRanges: document.orderedRanges,
                // definitionResults: document.definitionResults,
                // referenceResults: document.referenceResults,
                hoverResults: document.hoverResults,
                monikers: document.monikers,
                packageInformation: document.packageInformation,
            }),
        })
    }
    console.timeEnd('gathering documents')

    // const insertDefinitionsOrReferences = async (
    //     map: DefaultMap<Id, DefaultMap<Id, Set<Id>>>,
    //     inserter: TableInserter<DefinitionModel | ReferenceModel, new () => DefinitionModel | ReferenceModel>
    // ): Promise<void> => {
    //     for (const [documentId, rangeMonikers] of map) {
    //         for (const [rangeId, monikerIds] of rangeMonikers) {
    //             for (const monikerId of monikerIds) {
    //                 const range = assertDefined(rangeId, 'range', correlator.rangeData)
    //                 const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)
    //                 const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

    //                 await inserter.insert({
    //                     scheme: moniker.scheme,
    //                     identifier: moniker.identifier,
    //                     documentPath,
    //                     ...range,
    //                 })
    //             }
    //         }
    //     }
    // }

    // Insert definitions and references correlated by finalizing the
    // documents in the loop above. This will be used to search for ranges
    // by monikers.

    // TODO - can differentiate by a bool flag in same table? Save space?
    // console.time('inserting definitions')
    // await insertDefinitionsOrReferences(definitions, definitionInserter)
    // console.timeEnd('inserting definitions')

    // console.time('inserting references')
    // await insertDefinitionsOrReferences(references, referenceInserter)
    // console.timeEnd('inserting references')

    // TODO - tune
    const size = 50

    const chunks = new Array(size).fill(null).map(() => ({
        paths: new Map<Id, string>(),
        qualifiedRanges: new Map<Id, QualifiedRangeId[]>(),
    }))

    const chunkItems = (data: Map<Id, Map<Id, Id[]>>): void => {
        for (const [id, documentRanges] of data) {
            // Flatten map into list of qualified ranges
            let qualifiedRanges: QualifiedRangeId[] = []
            for (const [documentId, rangeIds] of documentRanges) {
                qualifiedRanges = qualifiedRanges.concat(rangeIds.map(rangeId => ({ documentId, rangeId })))
            }

            // Insert qualifiied ranges into chunk
            const chunk = chunks[hashKey(id, size)]
            chunk.qualifiedRanges.set(id, qualifiedRanges)

            for (const documentId of documentRanges.keys()) {
                // Add paths into the chunk where they are used
                chunk.paths.set(documentId, assertDefined(documentId, 'documentPath', correlator.documentPaths))
            }
        }
    }

    console.time('chunking items')
    chunkItems(correlator.definitionData)
    chunkItems(correlator.referenceData)
    console.timeEnd('chunking items')

    console.time('writing chunks')
    for (let i = 0; i < size; i++) {
        await chunkInserter.insert({
            id: i,
            data: await encodeJSON({
                paths: chunks[i].paths,
                qualifiedRanges: chunks[i].qualifiedRanges,
            }),
        })
    }

    console.timeEnd('writing chunks')

    // const writeMonikerLookupItems = async (
    //     inserter: TableInserter<DefinitionModel | ReferenceModel, new () => DefinitionModel | ReferenceModel>,
    //     monikers: Id[],
    //     documentRanges: Map<Id, Id[]>
    // ): Promise<void> => {
    //     for (const monikerId of monikers) {
    //         const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)

    //         for (const [documentId, rangeIds] of documentRanges) {
    //             const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

    //             for (const rangeId of rangeIds) {
    //                 const range = assertDefined(rangeId, 'range', correlator.rangeData)

    //                 await inserter.insert({
    //                     scheme: moniker.scheme,
    //                     identifier: moniker.identifier,
    //                     documentPath,
    //                     ...range,
    //                 })
    //             }
    //         }
    //     }
    // }

    console.time('correlating defs/refs')

    const definitionMonikers = new DefaultMap<Id, Set<Id>>(() => new Set<Id>())
    const referenceMonikers = new DefaultMap<Id, Set<Id>>(() => new Set<Id>())

    for (const range of correlator.rangeData.values()) {
        if (range.monikers.length === 0) {
            continue
        }

        if (range.definitionResult !== undefined) {
            const set = definitionMonikers.getOrDefault(range.definitionResult)
            for (const monikerId of range.monikers) {
                set.add(monikerId)
            }
        }

        if (range.referenceResult !== undefined) {
            const set = referenceMonikers.getOrDefault(range.referenceResult)
            for (const monikerId of range.monikers) {
                set.add(monikerId)
            }
        }
    }

    for (const [id, documentRanges] of correlator.definitionData) {
        for (const monikerId of definitionMonikers.getOrDefault(id)) {
            const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)

            for (const [documentId, rangeIds] of documentRanges) {
                const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

                for (const rangeId of rangeIds) {
                    const range = assertDefined(rangeId, 'range', correlator.rangeData)

                    await definitionInserter.insert({
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                        documentPath,
                        ...range,
                    })
                }
            }
        }
    }

    for (const [id, documentRanges] of correlator.referenceData) {
        for (const monikerId of referenceMonikers.getOrDefault(id)) {
            const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)

            for (const [documentId, rangeIds] of documentRanges) {
                const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

                for (const rangeId of rangeIds) {
                    const range = assertDefined(rangeId, 'range', correlator.rangeData)

                    await referenceInserter.insert({
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                        documentPath,
                        ...range,
                    })
                }
            }
        }
    }

    // for (const range of correlator.rangeData.values()) {
    //     if (range.monikers.length === 0) {
    //         continue
    //     }

    //     if (range.definitionResult) {
    //         await writeMonikerLookupItems(
    //             definitionInserter,
    //             range.monikers,
    //             assertDefined(range.definitionResult, 'definitionResult', correlator.definitionData)
    //         )
    //     }

    //     if (range.referenceResult) {
    //         await writeMonikerLookupItems(
    //             referenceInserter,
    //             range.monikers,
    //             assertDefined(range.referenceResult, 'referenceResult', correlator.referenceData)
    //         )
    //     }
    // }

    console.timeEnd('correlating defs/refs')

    // console.time('inserting definitions')
    // await insertDefinitionsOrReferences(definitions, definitionInserter)
    // console.timeEnd('inserting definitions')

    // console.time('inserting references')
    // await insertDefinitionsOrReferences(references, referenceInserter)
    // console.timeEnd('inserting references')

    // const getQualifiedRanges = (
    //     // map from docuemnt id to range ids
    //     documentRanges: Map<Id, Id[]>,
    //     // list of monikers on the current range
    //     monikerIds: Id[],
    //     // definition or references submap for the current document
    //     rangeMonikerMap: DefaultMap<Id, Set<Id>>
    // ): QualifiedRangeId[] => {
    //     let values: QualifiedRangeId[] = []
    //     for (const [documentId, rangeIds] of documentRanges) {
    //         // TODO
    //         document.paths.set(documentId, assertDefined(documentId, 'documentPath', correlator.documentPaths))
    //         if (rangeIds.length > 10000) {
    //             console.log(rangeIds.length)
    //         }
    //         // Resolve the "document" field from the "item" edge. This will correlate
    //         // the referenced range identifier with the document in which it belongs.
    //         // values = values.concat(rangeIds.map(id => ({ documentId, rangeId: id })))
    //         // TODO - re-enable!
    //         // if (documentId !== currentDocumentId) {
    //         //     continue
    //         // }
    //         // If this is results for the current document, construct the data that
    //         // will later be used to insert into the definition or reference table for
    //         // this document.
    //         // for (const id of rangeIds) {
    //         //     const monikerSet = rangeMonikerMap.getOrDefault(id)
    //         //     for (const monikerId of monikerIds) {
    //         //         monikerSet.add(monikerId)
    //         //     }
    //         // }
    //     }
    //     return values
    // }

    // Ensure all records are written
    await metaInserter.flush()
    await documentInserter.flush()
    await chunkInserter.flush()
    await definitionInserter.flush()
    await referenceInserter.flush()

    // Gather all package information that is referenced by an exported
    // moniker. These will be the packages that are provided by the repository
    // represented by this LSIF dump.

    const packageHashes: Package[] = []
    for (const monikerId of correlator.exportedMonikers) {
        const source = assertDefined(monikerId, 'moniker', correlator.monikerData)
        const packageInformationId = assertId(source.packageInformation)
        const packageInfo = assertDefined(packageInformationId, 'packageInformation', correlator.packageInformationData)

        packageHashes.push({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })
    }

    // Ensure packages are unique
    const exportedPackages = uniqWith(packageHashes, isEqual)

    // Gather all imporpted moniker identifiers along with their package
    // information. These will be the packages that are a dependency of the
    // repository represented by this LSIF dump.

    const packages = new Map<string, Package>()
    const packageIdentifiers = new DefaultMap<string, string[]>(() => [])
    for (const monikerId of correlator.importedMonikers) {
        const source = assertDefined(monikerId, 'moniker', correlator.monikerData)
        const packageInformationId = assertId(source.packageInformation)
        const packageInfo = assertDefined(packageInformationId, 'packageInformation', correlator.packageInformationData)

        const key = `${source.scheme}::${packageInfo.name}::${packageInfo.version}`
        packages.set(key, { scheme: source.scheme, name: packageInfo.name, version: packageInfo.version })
        packageIdentifiers.getOrDefault(key).push(source.identifier)
    }

    // Create a unique list of package information and imported symbol pairs.
    // Ensure that each pacakge is represented only once in the list.

    const importedReferences = Array.from(packages.keys()).map(key => ({
        package: assertDefined(key, 'package', packages),
        identifiers: assertDefined(key, 'packageIdentifier', packageIdentifiers),
    }))

    console.timeEnd('the whole shebang')

    // Kick back the xrepo data needed to be inserted into the correlation database
    return { packages: exportedPackages, references: importedReferences }
}

/**
 * Flatten the definition result, reference result, hover results, and monikers of range
 * and result set items by following next links in the graph. This needs to be run over
 * each range before committing them to a document.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param id The item identifier.
 * @param item The range or result set item.
 */
function canonicalizeItem(correlator: Correlator, id: Id, item: RangeData | ResultSetData): void {
    const monikers = new Set<Id>()
    if (item.monikers.length > 0) {
        // If we have any monikers attached to this item, then we only need to look at the
        // monikers reachable from any attached moniker. All other attached monikers are
        // necessarily reachable.

        for (const monikerId of reachableMonikers(correlator.monikerSets, item.monikers[0])) {
            if (assertDefined(monikerId, 'moniker', correlator.monikerData).kind !== MonikerKind.local) {
                monikers.add(monikerId)
            }
        }
    }

    const nextId = correlator.nextData.get(id)
    if (nextId !== undefined) {
        // If we have a next edge to a result set, get it and canonicalize it first. This
        // will recursively look at any result that that it can reach that hasn't yet been
        // canonicalized.

        const nextItem = assertDefined(nextId, 'resultSet', correlator.resultSetData)
        canonicalizeItem(correlator, nextId, nextItem)

        // Add each moniker of the next set to this item
        for (const monikerId of nextItem.monikers) {
            monikers.add(monikerId)
        }

        // If we do not have a definition, reference, or hover result, take the result
        // value from the next item.

        if (item.definitionResult === undefined) {
            item.definitionResult = nextItem.definitionResult
        }

        if (item.referenceResult === undefined) {
            item.referenceResult = nextItem.referenceResult
        }

        if (item.hoverResult === undefined) {
            item.hoverResult = nextItem.hoverResult
        }
    }

    // Update our moniker sets (our normalized sets and any monikers of our next item)
    item.monikers = Array.from(monikers)

    // Remove the next edge so we don't traverse it a second time
    correlator.nextData.delete(id)
}

/**
 * Create a self-contained document object from the data in the given correlator. This
 * method should also populate the definition and reference maps that are passed in.
 * They are initially empty.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param currentDocumentId The identifier of the document.
 * @param path The path of the document.
 * @param definitions A map from range identiifers to a set of moniker identifiers.
 * @param references A map from range identiifers to a set of moniker identifiers.
 */
function gatherDocument(
    correlator: Correlator,
    currentDocumentId: Id,
    path: string
    // definitions: DefaultMap<Id, Set<Id>>,
    // references: DefaultMap<Id, Set<Id>>
): DocumentData {
    const document = {
        path,
        ranges: new Map<Id, number>(),
        orderedRanges: [] as RangeData[],
        // definitionResults: new Map<Id, QualifiedRangeId[]>(),
        // referenceResults: new Map<Id, QualifiedRangeId[]>(),
        hoverResults: new Map<Id, string>(),
        monikers: new Map<Id, MonikerData>(),
        packageInformation: new Map<Id, PackageInformationData>(),
    }

    const addHover = (id: Id | undefined): void => {
        if (id === undefined || document.hoverResults.has(id)) {
            return
        }

        // Add hover result to the document, if defined and not a duplicate
        const data = assertDefined(id, 'hoverResult', correlator.hoverData)
        document.hoverResults.set(id, data)
    }

    const addPackageInformation = (id: Id | undefined): void => {
        if (id === undefined || document.packageInformation.has(id)) {
            return
        }

        // Add package information to the document, if defined and not a duplicate
        const data = assertDefined(id, 'packageInformation', correlator.packageInformationData)
        document.packageInformation.set(id, data)
    }

    const addMoniker = (id: Id | undefined): void => {
        if (id === undefined || document.monikers.has(id)) {
            return
        }

        // Add moniker to the document, if defined and not a duplicate
        const moniker = assertDefined(id, 'moniker', correlator.monikerData)
        document.monikers.set(id, moniker)

        // Add related package information to document
        addPackageInformation(moniker.packageInformation)
    }

    // const getQualifiedRanges = (
    //     // map from docuemnt id to range ids
    //     documentRanges: Map<Id, Id[]>,
    //     // list of monikers on the current range
    //     monikerIds: Id[],
    //     // definition or references submap for the current document
    //     rangeMonikerMap: DefaultMap<Id, Set<Id>>
    // ): QualifiedRangeId[] => {
    //     let values: QualifiedRangeId[] = []
    //     for (const [documentId, rangeIds] of documentRanges) {
    //         // TODO
    //         document.paths.set(documentId, assertDefined(documentId, 'documentPath', correlator.documentPaths))
    //         if (rangeIds.length > 10000) {
    //             console.log(rangeIds.length)
    //         }
    //         // Resolve the "document" field from the "item" edge. This will correlate
    //         // the referenced range identifier with the document in which it belongs.
    //         // values = values.concat(rangeIds.map(id => ({ documentId, rangeId: id })))
    //         // TODO - re-enable!
    //         // if (documentId !== currentDocumentId) {
    //         //     continue
    //         // }
    //         // If this is results for the current document, construct the data that
    //         // will later be used to insert into the definition or reference table for
    //         // this document.
    //         // for (const id of rangeIds) {
    //         //     const monikerSet = rangeMonikerMap.getOrDefault(id)
    //         //     for (const monikerId of monikerIds) {
    //         //         monikerSet.add(monikerId)
    //         //     }
    //         // }
    //     }
    //     return values
    // }

    const orderedRanges: (RangeData & { id: Id })[] = []
    for (const id of assertDefined(currentDocumentId, 'contains', correlator.containsData)) {
        const range = assertDefined(id, 'range', correlator.rangeData)
        orderedRanges.push({ id, ...range })

        addHover(range.hoverResult)

        for (const id of range.monikers) {
            addMoniker(id)
        }

        // if (range.definitionResult !== undefined) {
        //     const documentRanges = assertDefined(range.definitionResult, 'definitionResult', correlator.definitionData)
        //     const qualifiedRanges = getQualifiedRanges(documentRanges, range.monikers, definitions)
        //     document.definitionResults.set(range.definitionResult, qualifiedRanges)
        // }

        // if (range.referenceResult !== undefined) {
        //     const documentRanges = assertDefined(range.referenceResult, 'referenceResult', correlator.referenceData)
        //     const qualifiedRanges = getQualifiedRanges(documentRanges, range.monikers, references)
        //     document.referenceResults.set(range.referenceResult, qualifiedRanges)
        // }
    }

    // Sort ranges by their starting position
    orderedRanges.sort((a, b) => a.startLine - b.startLine || a.startCharacter - b.startCharacter)

    // Populate a reverse lookup so ranges can be queried by id
    // via `orderedRanges[range[id]]`.

    for (const [index, range] of orderedRanges.entries()) {
        document.ranges.set(range.id, index)
    }

    // eslint-disable-next-line require-atomic-updates
    document.orderedRanges = orderedRanges.map(({ id, ...range }) => range)

    return document
}

/**
 * Return the set of moniker identifiers which are reachable from the given value.
 * This relies on `monikerSets` being properly set up: each moniker edge `a -> b`
 * from the dump should ensure that `b` is a member of `monkerSets[a]`, and that
 * `a` is a member of `monikerSets[b]`.
 *
 * @param monikerSets A undirected graph of moniker ids.
 * @param id The initial moniker id.
 */
export function reachableMonikers(monikerSets: Map<Id, Set<Id>>, id: Id): Set<Id> {
    const monikerIds = new Set<Id>()
    let frontier = [id]

    while (frontier.length > 0) {
        const val = assertId(frontier.pop())
        if (monikerIds.has(val)) {
            continue
        }

        monikerIds.add(val)

        const nextValues = monikerSets.get(val)
        if (nextValues) {
            frontier = frontier.concat(Array.from(nextValues))
        }
    }

    // TODO - (efritz) should we sort these ids here instead of at query time?
    return monikerIds
}
