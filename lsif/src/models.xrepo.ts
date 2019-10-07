import { PrimaryGeneratedColumn, Column, Entity } from 'typeorm'
import { getBatchSize } from './util'
import { EncodedBloomFilter } from './encoding'

/**
 * The base class for `PackageModel` and `ReferenceModel` as they have nearly
 * identical column descriptions.
 */
class Package {
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
    @Column('text', { nullable: true })
    public version!: string | null

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
 * An entity within the cross-repo database. This maps a given repository and commit
 * pair to the package that it provides to other projects.
 */
@Entity({ name: 'packages' })
export class PackageModel extends Package {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = getBatchSize(5)
}

/**
 * An entity within the cross-repo database. This lists the dependencies of a given
 * repository and commit pair to support find global reference operations.
 */
@Entity({ name: 'references' })
export class ReferenceModel extends Package {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = getBatchSize(6)

    /**
     * A serialized bloom filter that encodes the set of symbols that this repository
     * and commit imports from the given package. Testing this filter will prevent
     * the backend from opening databases that will yield no results for a particular
     * symbol.
     *
     * The type of the column is determined by the database backend: we use Postgres
     * with type `bytea` in production, but use SQLite in unit tests with type `blob`.
     */
    @Column(process.env.JEST_WORKER_ID !== undefined ? 'blob' : 'bytea')
    public filter!: EncodedBloomFilter
}

/**
 * The entities composing the cross-repository database models.
 */
export const entities = [PackageModel, ReferenceModel]
