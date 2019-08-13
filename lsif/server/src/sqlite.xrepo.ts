import Sqlite from 'better-sqlite3'
import { BloomFilter } from 'bloomfilter'

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

export class CorrelationDatabase {
    private db!: Sqlite.Database

    public constructor(file: string) {
        this.db = new Sqlite(file)
    }

    public create() {
        this.db.exec(
            // Store external monikers with the package that repo/commit that contains them
            'Create Table packages (scheme string, name string, version string, repository string, "commit" string)'
        )

        // Store uses of external monikers, with the addition of a bloom filter that determines whether or not
        // the use of a particular URI is used within this package
        this.db.exec(
            'Create Table "references" (scheme string, name string, version string, repository string, "commit" string, filter string)'
        )
    }

    public close(): void {
        this.db.close()
    }

    public lookupRepositoryCommitByPackage(
        scheme: string,
        name: string,
        version: string
    ): { repository: string; commit: string } | undefined {
        const stmt = this.db.prepare(
            'Select p.* From packages p Where scheme = $scheme AND name = $name AND version = $version'
        )
        const result: PackageResult = stmt.get({ scheme, name, version })
        return result !== undefined ? { repository: result.repository, commit: result.commit } : undefined
    }

    public insertRepositoryCommitPackage(
        scheme: string,
        name: string,
        version: string,
        repository: string,
        commit: string
    ) {
        const stmt = this.db.prepare(
            'Insert Into packages (scheme, name, version, repository, "commit") Values ($scheme, $name, $version, $repository, $commit)'
        )
        stmt.run({ scheme, name, version, repository, commit })
    }

    public getAllRepositoryCommitReferences(
        scheme: string,
        name: string,
        version: string,
        uri: string
    ): { repository: string; commit: string }[] {
        const stmt = this.db.prepare(
            'Select r.* From "references" r Where scheme = $scheme AND name = $name AND version = $version'
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

        const stmt = this.db.prepare(
            'Insert Into "references" (scheme, name, version, repository, "commit", filter) Values ($scheme, $name, $version, $repository, $commit, $filter)'
        )
        stmt.run({ scheme, name, version, repository, commit, filter: buckets })
    }
}
