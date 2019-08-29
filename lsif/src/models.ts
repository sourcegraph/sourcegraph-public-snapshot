import { PrimaryGeneratedColumn, Column, Entity, PrimaryColumn, Index } from 'typeorm'

/**
 * An entity within the database describing LSIF data for a single repository
 * and commit pair. There should be only one metadata entity per database.
 */
@Entity({ name: 'meta' })
export class MetaModel {
    // A unique ID required by typeorm entities.
    @PrimaryGeneratedColumn()
    public id!: number

    // The version string of the input LSIF that created this database.
    @Column()
    public lsifVersion!: string

    // The internal version of the LSIF server that created this database.
    @Column()
    public sourcegraphVersion!: string
}

/**
 * An entity within the database describing LSIF data for a single repository and
 * commit pair. This contains a JSON-encoded `DocumentData` object that describes
 * relations within a single file. of the dump.
 */
@Entity({ name: 'documents' })
export class DocumentModel {
    // The root-relative path of the document.
    @PrimaryColumn()
    public uri!: string

    // The JSON-encoded document data.
    @Column()
    public value!: string
}

/**
 * The base class for `DefModel` and `RefModel` as they have identical column
 * descriptions.
 */
class Symbols {
    // A unique ID required by typeorm entities.
    @PrimaryColumn()
    public id!: number

    // The name of the package type (e.g. npm, pip).
    @Column()
    public scheme!: string

    // The unique identifier of the moniker.
    @Column()
    public identifier!: string

    // The uri of the document to which this reference belongs.
    @Column()
    public documentUri!: string

    // The zero-indexed line describing the start of this range.
    @Column()
    public startLine!: number

    // The zero-indexed line describing the end of this range.
    @Column()
    public endLine!: number

    // The zero-indexed line describing the start of this range.
    @Column()
    public startCharacter!: number

    // The zero-indexed line describing the end of this range.
    @Column()
    public endCharacter!: number
}

/**
 * An entity within the database describing LSIF data for a single repository and commit
 * pair. This maps external monikers to their range and the document that  contains the
 * definition of the moniker.
 */
@Entity({ name: 'defs' })
@Index(['scheme', 'identifier'])
export class DefModel extends Symbols {}

/**
 * An entity within the database describing LSIF data for a single repository and commit
 * pair. This maps imported monikers to their range and the document that contains a
 * reference to the moniker.
 */
@Entity({ name: 'refs' })
@Index(['scheme', 'identifier'])
export class RefModel extends Symbols {}

/**
 * An entity within the xrepo database. This maps a given repository and
 * commit pair to the package that it provides to other projects.
 */
@Entity({ name: 'packages' })
@Index(['scheme', 'name', 'version'])
export class PackageModel {
    // A unique ID required by typeorm entities.
    @PrimaryGeneratedColumn()
    public id!: number

    // The name of the package type (e.g. npm, pip).
    @Column()
    public scheme!: string

    // The name of the package this repository and commit provides.
    @Column()
    public name!: string

    // The version of the package this repository and commit provides.
    @Column()
    public version!: string

    // The name of the source repository.
    @Column()
    public repository!: string

    // The source commit.
    @Column()
    public commit!: string
}

/**
 * An entity within the xrepo database. This lists the dependencies of a given repository
 * and commit pair to support find global reference operations.
 */
@Entity({ name: 'references' })
@Index(['scheme', 'name', 'version'])
export class ReferenceModel {
    // A unique ID required by typeorm entities.
    @PrimaryGeneratedColumn()
    public id!: number

    // The name of the package type (e.g. npm, pip).
    @Column()
    public scheme!: string

    // The name of the package this repository and commit depends on.
    @Column()
    public name!: string

    // The version of the package this repository and commit depends on.
    @Column()
    public version!: string

    // The name of the source repository.
    @Column()
    public repository!: string

    // The source commit (revision hash).
    @Column()
    public commit!: string

    /**
     * A serialized bloom filter that encodes the set of symbols that this repository
     * and commcit imports from the given package. Testing this filter will prevent the
     * backend from opening databases that will yield no results for a particular symbol.
     */
    @Column()
    public filter!: string
}
