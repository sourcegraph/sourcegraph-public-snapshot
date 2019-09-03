import { PrimaryGeneratedColumn, Column, Entity, Index } from 'typeorm'

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
    @Column('text')
    public scheme!: string

    /**
     * The name of the package this repository and commit provides.
     */
    @Column('text')
    public name!: string

    /**
     * The version of the package this repository and commit provides.
     */
    @Column('text')
    public version!: string

    /**
     * The name of the source repository.
     */
    @Column('text')
    public repository!: string

    /**
     * The source commit.
     */
    @Column('text')
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
    @Column('text')
    public scheme!: string

    /**
     * The name of the package this repository and commit depends on.
     */
    @Column('text')
    public name!: string

    /**
     * The version of the package this repository and commit depends on.
     */
    @Column('text')
    public version!: string

    /**
     * The name of the source repository.
     */
    @Column('text')
    public repository!: string

    /**
     * The source commit (revision hash).
     */
    @Column('text')
    public commit!: string

    /**
     * A serialized bloom filter that encodes the set of symbols that this repository
     * and commit imports from the given package. Testing this filter will prevent the
     * backend from opening databases that will yield no results for a particular symbol.
     */
    @Column('text')
    public filter!: string
}
