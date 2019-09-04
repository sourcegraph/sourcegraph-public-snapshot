import { PrimaryGeneratedColumn, Column, Entity, PrimaryColumn, Index } from 'typeorm'

/**
 * An entity within the database describing LSIF data for a single repository
 * and commit pair. There should be only one metadata entity per database.
 */
@Entity({ name: 'meta' })
export class MetaModel {
    /**
     * A unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn('increment', { type: 'int' })
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
@Entity({ name: 'resultChunk' })
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
