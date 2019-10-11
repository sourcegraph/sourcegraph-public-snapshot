import { PrimaryGeneratedColumn, Column, Entity, PrimaryColumn, ViewEntity, ViewColumn } from 'typeorm'
import { getBatchSize } from './util'
import { EncodedBloomFilter } from './encoding'

/**
 * An entity within the cross-repo database. This tracks commit parentage and branch
 * heads for all known repositories.
 */
@Entity({ name: 'lsif_commits' })
export class Commit {
    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = getBatchSize(3)

    /**
     * A unique ID required by typeorm entities.
     */
    @PrimaryGeneratedColumn('increment', { type: 'int' })
    public id!: number

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

    /**
     * A parent commit. Multiple parents are represented by distinct rows
     * with the same `repository` and `commit`` fields. This value is an
     * empty string for a commit with no parent.
     */
    @Column('text', { name: 'parent_commit' })
    public parentCommit!: string
}

/**
 * An entity within the cross-repo database. A row with a repository and commit
 * indicates that there exists LSIF data for that pair.
 */
@Entity({ name: 'lsif_data_markers' })
export class LsifDataMarker {
    /**
     * The name of the source repository.
     */
    @PrimaryColumn('text')
    public repository!: string

    /**
     * The source commit.
     */
    @PrimaryColumn('text')
    public commit!: string

    /**
     * The number of model instances that can be inserted at once.
     */
    public static BatchSize = getBatchSize(2)
}

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
@Entity({ name: 'lsif_packages' })
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
@Entity({ name: 'lsif_references' })
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
 * A view that adds a `hasLsifData` column to the `lsif_commits` table. We define the
 * view in a migration as well as inline so that when we run unit tests (via a SQLite db)
 * we will also have access to the view.
 */
@ViewEntity({
    name: 'lsif_commits_with_lsif_data_markers',
    expression: `
        select
            c.repository,
            c."commit",
            c.parent_commit,
            exists (
                select 1
                from lsif_data_markers m
                where m.repository = c.repository and m."commit" = c."commit"
            ) as has_lsif_data
        from lsif_commits c
    `,
})
export class CommitWithLsifMarkers {
    @ViewColumn()
    public repository!: string

    @ViewColumn()
    public commit!: string

    @ViewColumn()
    public parentCommit!: string

    @ViewColumn()
    public hasLsifData!: boolean
}

/**
 * The entities composing the cross-repository database models.
 */
export const entities = [Commit, CommitWithLsifMarkers, LsifDataMarker, PackageModel, ReferenceModel]
