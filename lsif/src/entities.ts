import { Id, MonikerKind } from 'lsif-protocol';
import { Position } from 'vscode-languageserver-protocol';

/**
 * `DocumentData` stores data for a single document within an LSIF dump. The
 * data here can answer definitions, references, and hover queries if the
 * results are all contained within the same document.
 */
export interface DocumentData {
    /**
     * `ranges` is a mapping from range ID to the index of the range in the
     * `orderedRanges` array.
     */
    ranges: Map<Id, number>

    /**
     * `orderedRanges` is an array of range data sorted by startLine, then by
     * startCharacter. This allows us to perform binary search to find a
     * particular location subsumed by a range in the document.
     */
    orderedRanges: RangeData[]

    /**
     * `resultSets` map identifiers to a result set.
     */
    resultSets: Map<Id, ResultSetData>

    /**
     * `definitionResults` map identifiers to set of range identifiers that compose
     * the definition result.
     */
    definitionResults: Map<Id, Id[]>

    /**
     * `referenceResults` map identifiers to set of range identifiers that compose
     * the definition result.
     */
    referenceResults: Map<Id, Id[]>

    /**
     * `hovers` map identifiers to a hover result.
     */
    hovers: Map<Id, string>

    /**
     * `monikers` map identifiers to a moniker.
     */
    monikers: Map<Id, MonikerData>

    /**
     * `packageInformation` map identifiers to package information.
     */
    packageInformation: Map<Id, PackageInformationData>
}

/**
 * ~ResultObjectData` contains the set of fields shared by a range or a
 * result set vertex. It contains the same relevant edge data, which can
 * be subsequently queried in the containing document.
 */
interface ResultObjectData {
    /**
     * `monikers` is the set of moniker identifiers directly attached to
     * this range or result set. The moniker object can be queried by its
     * identifier within the containing document.
     */
    monikers: Id[]

    /**
     * `hoverResult` is the identifier of the hover result attached to this
     * range or result set, if one exists. The hover result object can be
     * queried by its identifier within the containing document.
     */
    hoverResult?: Id

    /**
     * `definitionResult` is the identifier of the definition result attached
     * to this range or result set, if one exists. The definition result object
     * can be queried by its identifier within the containing document.
     */
    definitionResult?: Id

    /**
     * `referenceResult` is the identifier of the reference result attached
     * to this range or result set, if one exists. The reference result object
     * can be queried by its identifier within the containing document.
     */
    referenceResult?: Id

    /**
     * `next` is the identifier of a result set attached to this range or result
     * set, if one exists. The result set object can be queried by its identifier
     * within the containing document.
     */
    next?: Id
}

/**
 * `RangeData` is an internal representation of a range vertex from an LSIF dump.
 * It contains the same relevant edge data, which can be subsequently queried in
 * the containing document.
 */
export interface RangeData extends ResultObjectData {
    /**
     * `start` is the start position of the range.
     */
    start: Position

    /**
     * `end` is the end position of the range.
     */
    end: Position
}

/**
 * `ResultSetData` is an internal representation of a result set vertex from an
 * LSIF dump. It contains the same relevant edge data, which can be subsequently
 * queried in the containing document.
 */
export interface ResultSetData extends ResultObjectData {}

/**
 * `MonikerData` holds data about a moniker attached to a range or a result set.
 */
export interface MonikerData {
    /**
     * `kind` is the kind of moniker (e.g. local, import, export).
     */
    kind: MonikerKind

    /**
     * `scheme` is the scheme of the moniker.
     */
    scheme: string

    /**
     * `identifier` is the identifier of the moniker.
     */
    identifier: string

    /**
     * `packageInformation` is the identifier of the package information to this
     * moniker, if one exists. The package information object can be queried by
     * its identifier within the containing document.
     */
    packageInformation?: Id
}

/**
 * `PackageInformationData` holds additional data about a non-local moniker.
 */
export interface PackageInformationData {
    /**
     * `name` is the name of the package the moniker describes.
     */
    name: string

    /**
     * `version` is the version of the package the moniker describes.
     */
    version: string
}

/**
 * `FlattenedRange` is an LSP range that has been squashed into a single layer.
 * This saves some space when serialized as JSON.
 */
export interface FlattenedRange {
    /**
     * `startLine` is the line on which the range starts (0-indexed, inclusive).
     */
    startLine: number

    /**
     * `startCharacter` is the line on which the range ends (0-indexed, inclusive).
     */
    startCharacter: number

    /**
     * `endLine` is the chracter on which the range starts (0-indexed, inclusive).
     */
    endLine: number

    /**
     * `endCharacter` is the chracter on which the range ends (0-indexed, inclusive).
     */
    endCharacter: number
}
