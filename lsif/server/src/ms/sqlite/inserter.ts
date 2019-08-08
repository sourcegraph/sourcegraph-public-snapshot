/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as Sqlite from 'better-sqlite3'

export class Inserter {
    private sqlStmt: Sqlite.Statement
    private batch: any[]

    public constructor(
        private db: Sqlite.Database,
        private stmt: string,
        private numberOfArgs: number,
        private batchSize: number
    ) {
        this.sqlStmt = this.makeStatement(this.batchSize)
        this.batch = []
    }

    public do(...params: any[]): void {
        if (params.length !== this.numberOfArgs) {
            throw new Error(`Wrong number of arguments. Expected ${this.numberOfArgs} but got ${params.length}`)
        }
        this.batch.push(...params)
        if (this.batch.length === this.numberOfArgs * this.batchSize) {
            this.sqlStmt.run(...this.batch)
            this.batch = []
        }
    }

    public finish(): void {
        if (this.batch.length === 0) {
            return
        }
        const finalStatement = this.makeStatement(this.batch.length / this.numberOfArgs)
        finalStatement.run(...this.batch)
    }

    private makeStatement(size: number): Sqlite.Statement {
        const args = `(${new Array(this.numberOfArgs).fill('?').join(',')})`
        return this.db.prepare(`${this.stmt} Values ${new Array(size).fill(args).join(',')}`)
    }
}
