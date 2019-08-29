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
    @Column('string')
    public lsifVersion!: string

    /**
     * The internal version of the LSIF server that created this database.
     */
    @Column('string')
    public sourcegraphVersion!: string
}

/**
 * An entity within the database describing LSIF data for a single repository and
 * commit pair. This contains a JSON-encoded `DocumentData` object that describes
 * relations within a single file. of the dump.
 */
@Entity({ name: 'documents' })
export class DocumentModel {
    /**
     * The root-relative path of the document.
     */
    @PrimaryColumn('string')
    public path!: string

    /**
     * The JSON-encoded document data.
     */
    @Column('string')
    public value!: string
}

/**
 * The base class for `DefModel` and `RefModel` as they have identical column
 * descriptions.
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
    @Column('string')
    public scheme!: string

    /**
     * The unique identifier of the moniker.
     */
    @Column('string')
    public identifier!: string

    /**
     * The path of the document to which this reference belongs.
     */
    @Column('string')
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
@Entity({ name: 'defs' })
@Index(['scheme', 'identifier'])
export class DefModel extends Symbols { }

/**
 * An entity within the database describing LSIF data for a single repository and commit
 * pair. This maps imported monikers to their range and the document that contains a
 * reference to the moniker.
 */
@Entity({ name: 'refs' })
@Index(['scheme', 'identifier'])
export class RefModel extends Symbols { }

/**
 * An entity within the xrepo database. This maps a given repository and
 * commit pair to the package that it provides to other projects.
 */
@Entity({ name: 'packages' })
@Index(['scheme', 'name', 'version'])
export class PackageModel {
    /**
     * A unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn('increment', { type: 'int' })
    public id!: number

    /**
     * The name of the package type (e.g. npm, pip).
     */
    @Column('string')
    public scheme!: string

    /**
     * The name of the package this repository and commit provides.
     */
    @Column('string')
    public name!: string

    /**
     * The version of the package this repository and commit provides.
     */
    @Column('string')
    public version!: string

    /**
     * The name of the source repository.
     */
    @Column('string')
    public repository!: string

    /**
     * The source commit.
     */
    @Column('string')
    public commit!: string
}

/**
 * An entity within the xrepo database. This lists the dependencies of a given repository
 * and commit pair to support find global reference operations.
 */
@Entity({ name: 'references' })
@Index(['scheme', 'name', 'version'])
export class ReferenceModel {
    /**
     * A unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn('increment', { type: 'int' })
    public id!: number

    /**
     * The name of the package type (e.g. npm, pip).
     */
    @Column('string')
    public scheme!: string

    /**
     * The name of the package this repository and commit depends on.
     */
    @Column('string')
    public name!: string

    /**
     * The version of the package this repository and commit depends on.
     */
    @Column('string')
    public version!: string

    /**
     * The name of the source repository.
     */
    @Column('string')
    public repository!: string

    /**
     * The source commit (revision hash).
     */
    @Column('string')
    public commit!: string

    /**
     * A serialized bloom filter that encodes the set of symbols that this repository
     * and commit imports from the given package. Testing this filter will prevent the
     * backend from opening databases that will yield no results for a particular symbol.
     */
    @Column('string')
    public filter!: string
}
