import Sqlite from 'better-sqlite3'

interface PackageResult {
    repository: string
    commit: string
}

export class CorrelationDatabase {
    private db!: Sqlite.Database

    public constructor(file: string) {
        this.db = new Sqlite(file)
    }

    public create() {
        this.db.exec(
            'Create Table packages (scheme string, name string, version string, repository string, "commit" string)'
        )
    }

    public close(): void {
        this.db.close()
    }

    public lookup(scheme: string, name: string, version: string): { repository: string; commit: string } | undefined {
        const stmt = this.db.prepare(
            'Select p.* From packages p WHERE scheme = $scheme AND name = $name AND version = $version'
        )
        const result: PackageResult = stmt.get({ scheme, name, version })
        return result !== undefined ? { repository: result.repository, commit: result.commit } : undefined
    }

    public insert(scheme: string, name: string, version: string, repository: string, commit: string) {
        const stmt = this.db.prepare(
            'Insert Into packages (scheme, name, version, repository, "commit") Values ($scheme, $name, $version, $repository, $commit)'
        )
        stmt.run({ scheme, name, version, repository, commit })
    }
}
