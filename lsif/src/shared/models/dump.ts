import * as lsif from 'lsif-protocol'
import { Column, Entity, Index, PrimaryColumn } from 'typeorm'
import { calcSqliteBatchSize } from './util'

export type DocumentId = lsif.Id
export type DocumentPath = string
export type RangeId = lsif.Id
export type DefinitionResultId = lsif.Id
export type ReferenceResultId = lsif.Id
export type DefinitionReferenceResultId = DefinitionResultId | ReferenceResultId
export type HoverResultId = lsif.Id
export type MonikerId = lsif.Id
export type PackageInformationId = lsif.Id

/**
 * A type that describes a gzipped and JSON-encoded value of type `T`.
 */
export type JSONEncoded<T> = Buffer

/**
 * A type of hashed value created by hashing a value of type `T` and performing
 * the modulus with a value of type `U`. This is to link the index of a result
 * chunk to the hashed value of the identifiers stored within it.
 */
export type HashMod<T, U> = number

/**
n entity within the database describing LSIF data for a single repository
 * and commit pair. There should be only one metadata entity per database.
 */
@Entity({ name: 'meta' })
export class MetaModel {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = calcSqliteBatchSize(4)

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
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = calcSqliteBatchSize(2)

    /**
     * The root-relative path of the document.
     */
    @PrimaryColumn('text')
    public path!: DocumentPath

    /**
     * The JSON-encoded document data.
     */
    @Column('blob')
    public data!: JSONEncoded<DocumentData>
}

/**
 * An entity within the database describing LSIF data for a single repository and
 * commit pair. This contains a JSON-encoded `ResultChunk` object that describes
 * a subset of the definition and reference results of the dump.
 */
@Entity({ name: 'resultChunks' })
export class ResultChunkModel {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = calcSqliteBatchSize(2)

    /**
     * The identifier of the chunk. This is also the index of the chunk during its
     * construction, and the identifiers contained in this chunk hash to this index
     * (modulo the total number of chunks for the dump).
     */
    @PrimaryColumn('int')
    public id!: HashMod<DefinitionReferenceResultId, MetaModel['numResultChunks']>

    /**
     * The JSON-encoded chunk data.
     */
    @Column('blob')
    public data!: JSONEncoded<ResultChunkData>
}

/**
 * The base class for `DefinitionModel` and `ReferenceModel` as they have identical
 * column descriptions.
 */
class Symbols {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = calcSqliteBatchSize(8)

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
    public documentPath!: DocumentPath

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
     * A mapping from range identifiers to range data.
     */
    ranges: Map<RangeId, RangeData>

    /**
     * A map of hover result identifiers to hover results normalized as a single
     * string.
     */
    hoverResults: Map<HoverResultId, string>

    /**
     * A map of moniker identifiers to moniker data.
     */
    monikers: Map<MonikerId, MonikerData>

    /**
     * A map of package information identifiers to package information data.
     */
    packageInformation: Map<PackageInformationId, PackageInformationData>
}

/**
 * A range identifier that also specifies the identifier of the document to
 * which it belongs. This is sometimes necessary as we hold definition and
 * reference results between packages, but the identifier of the range must be
 * looked up in a map of another encoded document.
 */
export interface DocumentIdRangeId {
    /**
     * The identifier of the document. The path of the document can be queried
     * by this identifier in the containing document.
     */
    documentId: DocumentId

    /**
     * The identifier of the range in the referenced document.
     */
    rangeId: RangeId
}

/**
 * A range identifier that also specifies the path of the document to which it
 * belongs. This is generally created by determining the path from an instance of
 * `DocumentIdRangeId`.
 */
export interface DocumentPathRangeId {
    /**
     * The path of the document.
     */
    documentPath: DocumentPath

    /**
     * The identifier of the range in the referenced document.
     */
    rangeId: RangeId
}

/**
 * A result chunk is a subset of the definition and reference result data for the
 * LSIF dump. Results are inserted into chunks based on the hash code of their
 * identifier (thus every chunk has a roughly proportional amount of data).
 */
export interface ResultChunkData {
    /**
     * A map from document identifiers to document paths. The document identifiers
     * in the `documentIdRangeIds` field reference a concrete path stored here.
     */
    documentPaths: Map<DocumentId, DocumentPath>

    /**
     * A map from definition or reference result identifiers to the ranges that
     * compose the result set. Each range is paired with the identifier of the
     * document in which it can be found.
     */
    documentIdRangeIds: Map<DefinitionReferenceResultId, DocumentIdRangeId[]>
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
     * The definition result object can be queried by its identifier within the containing
     * document.
     */
    definitionResultId?: DefinitionResultId

    /**
     * The identifier of the reference result attached to this range, if one exists.
     * The reference result object can be queried by its identifier within the containing
     * document.
     */
    referenceResultId?: ReferenceResultId

    /**
     * The identifier of the hover result attached to this range, if one exists. The
     * hover result object can be queried by its identifier within the containing
     * document.
     */
    hoverResultId?: HoverResultId

    /**
     * The set of moniker identifiers directly attached to this range. The moniker
     * object can be queried by its identifier within the
     * containing document.
     */
    monikerIds: Set<MonikerId>
}

/**
 * Data about a moniker attached to a range.
 */
export interface MonikerData {
    /**
     * The kind of moniker (e.g. local, import, export).
     */
    kind: lsif.MonikerKind

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
    packageInformationId?: PackageInformationId
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
    version: string | null
}

/**
 * The entities composing the database models.
 */
export const entities = [DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel]
