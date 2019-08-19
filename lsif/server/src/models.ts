import { PrimaryGeneratedColumn, Column, Entity, PrimaryColumn, Index } from 'typeorm'

/* eslint-disable @typescript-eslint/no-unused-vars */

/**
 * `MetaModel` is an entity within the database describing LSIF data for a single repository
 * and commit pair. There should be only one metadata entity per database.
 */
@Entity({ name: 'meta' })
export class MetaModel {
    /**
     * `id` is a unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn()
    public id!: number

    /**
     * `lsifVersion` is the version string of the input LSIF that created this database.
     */
    @Column()
    public lsifVersion!: string

    /**
     * `sourcegraphVersion` is the internal version of the LSIF server that created this database.
     */
    @Column()
    public sourcegraphVersion!: string
}

/**
 * `DocumentModel` is an entity within the database describing LSIF data for a single
 * repository and commit pair. This contains a JSON-encoded `DocumentData` object that
 * describes relations within a single file. of the dump.
 */
@Entity({ name: 'documents' })
export class DocumentModel {
    /**
     * `uri` is the root-relative path of the document.
     */
    @PrimaryColumn()
    public uri!: string

    /**
     * `value` is the JSON-encoded document data.
     */
    @Column()
    public value!: string
}

/**
 * `DefModel` is an entity within the database describing LSIF data for a single repository
 * and commit pair. This maps external monikers to their range and the document that contains
 * the definition of the moniker.
 */
@Entity({ name: 'defs' })
@Index(['scheme', 'identifier'])
export class DefModel {
    /**
     * `id` is a unique ID required by typeorm entities.
     */
    @PrimaryColumn()
    public id!: number

    /**
     * `scheme` describes the package manager type (e.g. npm, pip).
     */
    @Column()
    public scheme!: string

    /**
     * `identifier` describes the moniker.
     */
    @Column()
    public identifier!: string

    /**
     * `documentUri` is the uri of the document to which this definition belongs.
     */
    @Column()
    public documentUri!: string

    /**
     * `startLine` is the zero-indexed line describing the start of this range.
     */
    @Column()
    public startLine!: number

    /**
     * `endLine` is the zero-indexed line describing the end of this range.
     */
    @Column()
    public endLine!: number

    /**
     * `startCharacter` is the zero-indexed line describing the start of this range.
     */
    @Column()
    public startCharacter!: number

    /**
     * `endCharacter` is the zero-indexed line describing the end of this range.
     */
    @Column()
    public endCharacter!: number
}

/**
 * `RefModel` is an entity within the database describing LSIF data for a single repository
 * and commit pair. This maps imported monikers to their range and the document that contains
 * a reference to the moniker.
 */
@Entity({ name: 'refs' })
@Index(['scheme', 'identifier'])
export class RefModel {
    /**
     * `id` is a unique ID required by typeorm entities.
     */
    @PrimaryColumn()
    public id!: number

    /**
     * `scheme` describes the package manager type (e.g. npm, pip).
     */
    @Column()
    public scheme!: string

    /**
     * `identifier` describes the moniker.
     */
    @Column()
    public identifier!: string

    /**
     * `documentUri` is the uri of the document to which this reference belongs.
     */
    @Column()
    public documentUri!: string

    /**
     * `startLine` is the zero-indexed line describing the start of this range.
     */
    @Column()
    public startLine!: number

    /**
     * `endLine` is the zero-indexed line describing the end of this range.
     */
    @Column()
    public endLine!: number

    /**
     * `startCharacter` is the zero-indexed line describing the start of this range.
     */
    @Column()
    public startCharacter!: number

    /**
     * `endCharacter` is the zero-indexed line describing the end of this range.
     */
    @Column()
    public endCharacter!: number
}

/**
 * `PackageModel` is an entity within the xrepo database. This maps a given repository and
 * commit pair to the package that it provides to other projects.
 */
@Entity({ name: 'packages' })
@Index(['scheme', 'name', 'version'])
export class PackageModel {
    /**
     * `id` is a unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn()
    public id!: number

    /**
     * `scheme` describes the package manager type (e.g. npm, pip).
     */
    @Column()
    public scheme!: string

    /**
     * `name` is the name of the package this repository and commit provides.
     */
    @Column()
    public name!: string

    /**
     * `version` is the version of the package this repository and commit provides.
     */
    @Column()
    public version!: string

    /**
     * `repository` is the source repository.
     */
    @Column()
    public repository!: string

    /**
     * `commit` is the source commit.
     */
    @Column()
    public commit!: string
}

/**
 * `ReferenceModel` is an entity within the xrepo database. This lists the dependencies
 * of a given repository and commit pair to support find global reference operations.
 */
@Entity({ name: 'references' })
@Index(['scheme', 'name', 'version'])
export class ReferenceModel {
    /**
     * `id` is a unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn()
    public id!: number

    /**
     * `scheme` describes the package manager type (e.g. npm, pip).
     */
    @Column()
    public scheme!: string

    /**
     * `name` is the name of the package this repository and commit depends on.
     */
    @Column()
    public name!: string

    /**
     * `version` is the version of the package this repository and commit depends on.
     */
    @Column()
    public version!: string

    /**
     * `repository` is the source repository.
     */
    @Column()
    public repository!: string

    /**
     * `commit` is the source commit.
     */
    @Column()
    public commit!: string

    /**
     * `filter` is an serialized bloom filter that encodes the set of symbols that
     * this repository and commcit imports from the given package. Testing this
     * filter will prevent the backend from opening databases that will yield no
     * results for a particular symbol.
     */
    @Column()
    public filter!: string
}
