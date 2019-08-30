import { Id, MonikerKind } from 'lsif-protocol'

/**
 * Data for a single document within an LSIF dump. The data here can answer definitions,
 * references, and hover queries if the results are all contained within the same document.
 */
export interface DocumentData {
    /**
     * A mapping from range identifiers to the index of the range in the
     * `orderedRanges` array. We keep a mapping so we can look range data by
     * identifier quickly, and keep them sorted so we can find the range that
     * encloses a position quickly.
     */
    ranges: Map<Id, number>

    /**
     * An array of range data sorted by startLine, then by startCharacter. This
     * allows us to perform binary search to find a particular location subsumed
     * by a range in the document.
     */
    orderedRanges: RangeData[]

    /**
     * A map of result set identifiers to result set data. Result sets are like
     * ranges, but do  not have extents in the document.
     */
    resultSets: Map<Id, ResultSetData>

    /**
     * A map of definition result identifiers to a list of ids that compose the
     * definition result. Each id is paired with a document path, as result sets
     * can be shared between documents (necessitating cross-document queries).
     */
    definitionResults: Map<Id, {documentPath: string, id: Id}[]>

    /**
     ** A map of reference result identifiers to a list of ids that compose the
     * reference result. Each id is paired with a document path, as result sets
     * can be shared between documents (necessitating cross-document queries).
     */
    referenceResults: Map<Id, {documentPath: string, id: Id}[]>

    /**
     * A map of hover identifiers to hover results normalized as a single string.
     */
    hovers: Map<Id, string>

    /**
     * A map of moniker identifiers to moniker data.
     */
    monikers: Map<Id, MonikerData>

    /**
     * A map of package information identifiers to package information data.
     */
    packageInformation: Map<Id, PackageInformationData>
}

/**
 * The set of fields shared by a range or a result set vertex. It contains
 * the same relevant edge data, which can be subsequently queried in the
 * containing document.
 */
interface ResultObjectData {
    /**
     * The set of moniker identifiers directly attached to this range or result
     * set. The moniker object can be queried by its identifier within the
     * containing document.
     */
    monikers: Id[]

    /**
     * The identifier of the hover result attached to this range or result set,
     * if one exists. The hover result object can be queried by its identifier
     * within the containing document.
     */
    hoverResult?: Id

    /**
     * The identifier of the definition result attached to this range or result
     * set, if one exists. The definition result object can be queried by its
     * identifier within the containing document.
     */
    definitionResult?: Id

    /**
     * The identifier of the reference result attached to this range or result
     * set, if one exists. The reference result object can be queried by its
     * identifier within the containing document.
     */
    referenceResult?: Id

    /**
     * The identifier of a result set attached to this range or result set, if one
     * exists. The result set object can be queried by its identifier within the
     * containing document.
     */
    next?: Id
}

/**
 * An internal representation of a range vertex from an LSIF dump. It contains the same
 * relevant edge data, which can be subsequently queried in the containing document.
 */
export interface RangeData extends ResultObjectData, FlattenedRange {}

/**
 * An internal representation of a result set vertex from an LSIF dump. It contains the
 * same relevant edge data, which can be subsequently queried in the containing document.
 */
export interface ResultSetData extends ResultObjectData {}

/**
 * Data about a moniker attached to a range or a result set.
 */
export interface MonikerData {
    /**
     * The kind of moniker (e.g. local, import, export).
     */
    kind: MonikerKind

    /**
     * The name of the package type (e.g. npm, pip).
     */
    scheme: string

    /**
     * The unique identifier of the moniker.
     */
    identifier: string

    /**
     * The identifier of the package information to this moniker, if one exists.
     * The package information object can be queried by its identifier within the
     * containing document.
     */
    packageInformation?: Id
}

/**
 * Additional data about a non-local moniker.
 */
export interface PackageInformationData {
    /**
     * The name of the package the moniker describes.
     */
    name: string

    /**
     * The version of the package the moniker describes.
     */
    version: string
}

/**
 * An LSP range that has been squashed into a single layer.
 */
export interface FlattenedRange {
    /**
     * The line on which the range starts (0-indexed, inclusive).
     */
    startLine: number

    /**
     * The line on which the range ends (0-indexed, inclusive).
     */
    startCharacter: number

    /**
     * The character on which the range starts (0-indexed, inclusive).
     */
    endLine: number

    /**
     * The character on which the range ends (0-indexed, inclusive).
     */
    endCharacter: number
}
