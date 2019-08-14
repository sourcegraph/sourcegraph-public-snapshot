import Sqlite from 'better-sqlite3'

export interface SchemaTable {
    columns: SchemaColumn[]
    batchInsertSize: number
}

export interface SchemaColumn {
    name: string
    type: string
    indexed: boolean
}

export function createTables(db: Sqlite.Database, tables: { [K: string]: SchemaTable }): void {
    for (const table of Object.keys(tables)) {
        const declaredColumns = tables[table].columns.map(c => `"${c.name}" ${c.type}`)
        db.exec(`create table "${table}" (${declaredColumns.join(', ')})`)
    }
}

export function createIndices(db: Sqlite.Database, tables: { [K: string]: SchemaTable }): void {
    for (const table of Object.keys(tables)) {
        const indexedNames = tables[table].columns.filter(c => c.indexed).map(c => c.name)
        if (indexedNames.length > 0) {
            db.exec(`create index _${table} on "${table}" (${indexedNames.join(', ')})`)
        }
    }
}

export class Inserter {
    private statement: Sqlite.Statement
    private params: any[] = []

    public constructor(private db: Sqlite.Database, private name: string, private table: SchemaTable) {
        this.statement = this.makeStatement(this.table.batchInsertSize)
    }

    public do(...params: any[]): void {
        this.params.push(...params)

        if (this.params.length >= this.table.columns.length * this.table.batchInsertSize) {
            this.statement.run(...this.params)
            this.params = []
        }
    }

    public finish(): void {
        if (this.params.length > 0) {
            this.makeStatement(this.params.length / this.table.columns.length).run(...this.params)
        }
    }

    private makeStatement(size: number): Sqlite.Statement {
        const columns = this.table.columns.map(c => `"${c.name}"`)
        const placeholders = new Array(this.table.columns.length).fill('?').join(',')
        const batchPlaceholders = new Array(size).fill(`(${placeholders})`).join(',')
        return this.db.prepare(`insert into "${this.name}" (${columns}) values ${batchPlaceholders}`)
    }
}
