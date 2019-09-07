import { Column, Entity, Index, PrimaryColumn } from 'typeorm'
import { Id, MonikerKind } from 'lsif-protocol'

/**
 * An entity within the database describing LSIF data for a single repository
 * and commit pair. There should be only one metadata entity per database.
 */
@Entity({ name: 'meta' })
export class MetaModel {
    /**
     * A unique ID required by typeorm entities: always zero here.
     */
    @PrimaryColumn('int')
    public id!: number

    /**
     * The version string of the input LSIF that created this database.
     */
    @Column('text')
    public lsifVersion!: string

    /**
     * The internal version of the LSIF server that created this database.
     */
    @Column('text')
    public sourcegraphVersion!: string

    /**
     * The number of result chunks allocated when converting the dump stored
     * in this database. This is used as an upper bound for the hash into the
     * `resultChunks` table and must be record to keep the hash generation
     * stable.
     */
    @Column('int')
    public numResultChunks!: number
}

/**
 * An entity within the database describing LSIF data for a single repository and
 * commit pair. This contains a JSON-encoded `DocumentData` object that describes
 * relations within a single file of the dump.
 */
@Entity({ name: 'documents' })
export class DocumentModel {
    /**
     * The root-relative path of the document.
     */
    @PrimaryColumn('text')
    public path!: string

    /**
     * The JSON-encoded document data.
     */
    @Column('text')
    public data!: string
}

/**
 * An entity within the database describing LSIF data for a single repository and
 * commit pair. This contains a JSON-encoded `ResultChunk` object that describes
 * a subset of the definition and reference results of the dump.
 */
@Entity({ name: 'resultChunks' })
export class ResultChunkModel {
    /**
     * The identifier of the chunk. This is also the index of the chunk during its
     * construction, and the identifiers contained in this chunk hash to this index
     * (modulo the total number of chunks for the dump).
     */
    @PrimaryColumn('int')
    public id!: number

    /**
     * The JSON-encoded chunk data.
     */
    @Column('text')
    public data!: string
}

/**
 * The base class for `DefinitionModel` and `ReferenceModel` as they have identical
 * column descriptions.
 */
class Symbols {
    /**
     * A unique ID required by typeorm entities.
     */
    @PrimaryColumn('int')
    public id!: number

    /**
     * The name of the package type (e.g. npm, pip).
     */
    @Column('text')
    public scheme!: string

    /**
     * The unique identifier of the moniker.
     */
    @Column('text')
    public identifier!: string

    /**
     * The path of the document to which this reference belongs.
     */
    @Column('text')
    public documentPath!: string

    /**
     * The zero-indexed line describing the start of this range.
     */
    @Column('int')
    public startLine!: number

    /**
     * The zero-indexed line describing the end of this range.
     */
    @Column('int')
    public endLine!: number

    /**
     * The zero-indexed line describing the start of this range.
     */
    @Column('int')
    public startCharacter!: number

    /**
     * The zero-indexed line describing the end of this range.
     */
    @Column('int')
    public endCharacter!: number
}

/**
 * An entity within the database describing LSIF data for a single repository and commit
 * pair. This maps external monikers to their range and the document that  contains the
 * definition of the moniker.
 */
@Entity({ name: 'definitions' })
@Index(['scheme', 'identifier'])
export class DefinitionModel extends Symbols {}

/**
 * An entity within the database describing LSIF data for a single repository and commit
 * pair. This maps imported monikers to their range and the document that contains a
 * reference to the moniker.
 */
@Entity({ name: 'references' })
@Index(['scheme', 'identifier'])
export class ReferenceModel extends Symbols {}

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
     * A map of hover result identifiers to hover results normalized as a single
     * string.
     */
    hoverResults: Map<Id, string>

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
 * A range identifier that also specifies the path of the document to which it
 * belongs. This is sometimes necessary as we hold definition and refererence
 * results between packages, but the identifier of the range must be looked up
 * in a map of another encoded document.
 */
export interface QualifiedRangeId {
    /**
     * The identifier of the document. The path of the document can be queried
     * by this identifier in the containing document.
     */
    documentId: Id

    /**
     * The identifier of the range in the referenced document.
     */
    rangeId: Id
}

/**
 * A result chunk is a subset of the definition and reference result data for the
 * LSIF dump. Results are inserted into chunks based on the hash code of their
 * identifier (thus every chunk has a roughly proportional amount of data).
 */
export interface ResultChunkData {
    /**
     * A map from document identifiers to document paths. The document identifiers
     * in the qualified ranges map reference a concrete path stored here.
     */
    paths: Map<Id, string>

    /**
     * A map from definition or reference result identifiers to the qualified ranges
     * that compose the result set.
     */
    qualifiedRanges: Map<Id, QualifiedRangeId[]>

    // TODO - suffix like things with Id
}

/**
 * An internal representation of a range vertex from an LSIF dump. It contains the same
 * relevant edge data, which can be subsequently queried in the containing document. The
 * data that was reachable via a result set has been collapsed into this object during
 * import.
 */
export interface RangeData {
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

    /**
     * The identifier of the definition result attached to this range, if one exists.
     * The definition result object can be queried by its * identifier within the containing
     * document.
     */
    definitionResult?: Id

    /**
     * The identifier of the reference result attached to this range, if one exists.
     * The reference result object can be queried by its identifier within the containing
     * document.
     */
    referenceResult?: Id

    /**
     * The identifier of the hover result attached to this range, if one exists. The
     * hover result object can be queried by its identifier within the containing
     * document.
     */
    hoverResult?: Id

    /**
     * The set of moniker identifiers directly attached to this range. The moniker
     * object can be queried by its identifier within the
     * containing document.
     */
    monikers: Id[]
}

/**
 * Data about a moniker attached to a range.
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
