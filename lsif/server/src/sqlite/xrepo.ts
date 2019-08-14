import Sqlite from 'better-sqlite3'
import { BloomFilter } from 'bloomfilter'
import { SchemaTable, createTables, createIndices, Inserter } from './sqlite'

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter
const BLOOM_FILTER_BITS = 64 * 1024
const BLOOM_FILTER_NUM_HASH_FUNCTIONS = 16

interface PackageResult {
    repository: string
    commit: string
}

interface ReferenceResult {
    repository: string
    commit: string
    filter: string
}

const SCHEMA_TABLES: { [K: string]: SchemaTable } = {
    // Store external monikers with the package that repo/commit that contains them
    packages: {
        columns: [
            { name: 'scheme', type: 'string', indexed: false },
            { name: 'name', type: 'string', indexed: false },
            { name: 'version', type: 'string', indexed: false },
            { name: 'repository', type: 'string', indexed: false },
            { name: 'commit', type: ' string', indexed: false },
        ],
        batchInsertSize: 1,
    },

    // Store uses of external monikers, with the addition of a bloom filter that determines whether or not
    // the use of a particular URI is used within this package
    references: {
        columns: [
            { name: 'scheme', type: 'string', indexed: false },
            { name: 'name', type: 'string', indexed: false },
            { name: 'version', type: 'string', indexed: false },
            { name: 'repository', type: 'string', indexed: false },
            { name: 'commit', type: 'string', indexed: false },
            { name: 'filter', type: 'string', indexed: false },
        ],
        batchInsertSize: 1,
    },
}

export class CorrelationDatabase {
    private db!: Sqlite.Database
    private packageInserter: Inserter
    private referenceInserter: Inserter

    public constructor(file: string) {
        this.db = new Sqlite(file)
        this.packageInserter = new Inserter(this.db, 'packages', SCHEMA_TABLES['packages'])
        this.referenceInserter = new Inserter(this.db, 'references', SCHEMA_TABLES['references'])
    }

    public create() {
        createTables(this.db, SCHEMA_TABLES)
        createIndices(this.db, SCHEMA_TABLES)
    }

    public close(): void {
        this.db.close()
    }

    //
    // Querying

    public lookupRepositoryCommitByPackage(
        scheme: string,
        name: string,
        version: string
    ): { repository: string; commit: string } | undefined {
        const stmt = this.db.prepare(
            'select p.* from "packages" p where "scheme" = $scheme and "name" = $name and "version" = $version'
        )
        const result: PackageResult = stmt.get({ scheme, name, version })
        return result !== undefined ? { repository: result.repository, commit: result.commit } : undefined
    }

    public getAllRepositoryCommitReferences(
        scheme: string,
        name: string,
        version: string,
        uri: string
    ): { repository: string; commit: string }[] {
        const stmt = this.db.prepare(
            'select r.* from "references" r where "scheme" = $scheme and "name" = $name and "version" = $version'
        )

        const results: ReferenceResult[] = stmt.all({ scheme, name, version })

        const filtered = []
        for (const result of results) {
            // TODO(efritz) - decode smarter
            const filter = new BloomFilter(JSON.parse(result.filter), BLOOM_FILTER_NUM_HASH_FUNCTIONS)

            if (filter.test(uri)) {
                filtered.push({
                    repository: result.repository,
                    commit: result.commit,
                })
            }
        }

        return filtered
    }

    //
    // Insertion

    public insertRepositoryCommitPackage(
        scheme: string,
        name: string,
        version: string,
        repository: string,
        commit: string
    ) {
        this.packageInserter.do(scheme, name, version, repository, commit)
    }

    public insertRepositoryCommitReference(
        scheme: string,
        name: string,
        version: string,
        repository: string,
        commit: string,
        uris: string[]
    ): void {
        const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
        uris.forEach(uri => filter.add(uri))

        // TODO(efritz) - encode smarter
        const buckets = JSON.stringify([].slice.call(filter.buckets))

        this.referenceInserter.do(scheme, name, version, repository, commit, buckets)
    }
}
