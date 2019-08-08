/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as fs from 'fs'
import Sqlite from 'better-sqlite3'
import {
    Edge,
    Vertex,
    ElementTypes,
    VertexLabels,
    Document,
    Range,
    Project,
    MetaData,
    EdgeLabels,
    contains,
    PackageInformation,
    item,
} from 'lsif-protocol'
import { itemPropertyShortForms } from './compress'
import { Inserter } from './inserter'

export class GraphStore {
    private db: Sqlite.Database
    private insertContentStmt: Sqlite.Statement
    private vertexInserter: Inserter
    private edgeInserter: Inserter
    private itemInserter: Inserter
    private rangeInserter: Inserter
    private documentInserter: Inserter
    private pendingRanges: Map<number | string, Range>

    public constructor(
        filename: string,
        private stringify: (element: Vertex | Edge) => string,
        private shortForm: (element: Vertex | Edge) => number
    ) {
        this.pendingRanges = new Map()
        try {
            fs.unlinkSync(filename)
        } catch (err) {}
        this.db = new Sqlite(filename)
        this.db.pragma('synchronous = OFF')
        this.db.pragma('journal_mode = MEMORY')
        this.createTables()
        this.db.exec(`Insert into format (format) Values ('graph')`)

        this.vertexInserter = new Inserter(this.db, 'Insert Into vertices (id, label, value)', 3, 128)
        this.rangeInserter = new Inserter(
            this.db,
            'Insert into ranges (id, belongsTo, startLine, startCharacter, endLine, endCharacter)',
            6,
            128
        )
        this.documentInserter = new Inserter(this.db, 'Insert Into documents (uri, id)', 2, 5)
        this.insertContentStmt = this.db.prepare('Insert Into contents (id, content) VALUES (?, ?)')

        this.edgeInserter = new Inserter(this.db, 'Insert Into edges (id, label, outV, inV)', 4, 128)
        this.itemInserter = new Inserter(this.db, 'Insert Into items (id, outV, inV, document, property)', 5, 128)
    }

    private createTables(): void {
        // Vertex information
        this.db.exec('Create Table format (format Text Not Null)')
        this.db.exec(
            'Create Table vertices (id Integer Unique Primary Key, label Integer Not Null, value Text Not Null)'
        )
        this.db.exec('Create Table meta (id Integer Unique Primary Key, value Text Not Null)')
        this.db.exec(
            'Create Table ranges (id Integer Unique Primary Key, belongsTo Integer Not Null, startLine Integer Not Null, startCharacter Integer Not Null, endLine Integer Not Null, endCharacter Integer Not Null)'
        )
        this.db.exec('Create Table documents (uri Text Unique Primary Key, id Integer Not Null)')
        this.db.exec('Create Table contents (id Integer Unique Primary Key, content Blob Not Null)')

        // Edge information
        this.db.exec(
            'Create Table edges (id Integer Not Null, label Integer Not Null, outV Integer Not Null, inV Integer Not Null)'
        )
        this.db.exec(
            'Create Table items (id Integer Not Null, outV Integer Not Null, inV Integer Not Null, document Integer Not Null, property Integer)'
        )
    }

    private createIndices(): void {
        // Index label, outV and inV on edges
        this.db.exec('Create Index edges_outv on edges (outV, label)')
        this.db.exec('Create Index edges_inv on edges (inV, label)')
        this.db.exec(
            'Create Index ranges_index on ranges (belongsTo, startLine, endLine, startCharacter, endCharacter)'
        )
        this.db.exec('Create Index items_outv on items (outV)')
        this.db.exec('Create Index items_inv on items (inV)')
    }

    public runInsertTransaction(cb: (db: GraphStore) => void): void {
        this.db.transaction(() => {
            cb(this)
        })()
    }

    public insert(element: Edge | Vertex): void {
        if (element.type === ElementTypes.vertex) {
            switch (element.label) {
                case VertexLabels.metaData:
                    this.insertMetaData(element)
                    break
                case VertexLabels.project:
                    this.insertProject(element)
                    break
                case VertexLabels.document:
                    this.insertDocument(element)
                    break
                case VertexLabels.packageInformation:
                    this.insertPackageInformation(element)
                    break
                case VertexLabels.range:
                    this.insertRange(element)
                    break
                default:
                    this.insertVertex(element)
            }
        } else {
            switch (element.label) {
                case EdgeLabels.contains:
                    this.insertContains(element)
                    break
                case EdgeLabels.item:
                    this.insertItem(element)
                    break
                default:
                    this.insertEdge(element)
            }
        }
    }

    private insertVertex(vertex: Vertex): void {
        let value = this.stringify(vertex)
        let label = this.shortForm(vertex)
        this.vertexInserter.do(vertex.id, label, value)
    }

    private insertMetaData(vertex: MetaData): void {
        let value = this.stringify(vertex)
        this.db.exec(`Insert Into meta (id, value) Values (${vertex.id}, '${value}')`)
    }

    private insertContent(vertex: Document | Project | PackageInformation): void {
        if (vertex.contents === undefined || vertex.contents === null) {
            return
        }
        let contents = Buffer.from(vertex.contents, 'base64').toString('utf8')
        this.insertContentStmt.run(vertex.id, contents)
    }

    private insertProject(project: Project): void {
        if (project.resource !== undefined && project.contents !== undefined) {
            this.documentInserter.do(project.resource, project.id)
            this.insertContent(project)
        }
        let newProject = Object.assign(Object.create(null) as object, project)
        newProject.contents = undefined
        this.insertVertex(newProject)
    }

    private insertDocument(document: Document): void {
        this.documentInserter.do(document.uri, document.id)
        this.insertContent(document)
        let newDocument = Object.assign(Object.create(null) as object, document)
        newDocument.contents = undefined
        this.insertVertex(newDocument)
    }

    private insertPackageInformation(info: PackageInformation): void {
        if (info.uri !== undefined && info.contents !== undefined) {
            this.documentInserter.do(info.uri, info.id)
            this.insertContent(info)
        }
        let newInfo = Object.assign(Object.create(null) as object, info)
        newInfo.contents = undefined
        this.insertVertex(newInfo)
    }

    private insertRange(range: Range): void {
        this.insertVertex(range)
        this.pendingRanges.set(range.id, range)
    }

    private insertEdge(edge: Edge): void {
        let label = this.shortForm(edge)
        if (Edge.is11(edge)) {
            this.edgeInserter.do(edge.id, label, edge.outV, edge.inV)
        } else if (Edge.is1N(edge)) {
            for (let inV of edge.inVs) {
                this.edgeInserter.do(edge.id, label, edge.outV, inV)
            }
        }
    }

    private insertContains(contains: contains): void {
        let label = this.shortForm(contains)
        for (let inV of contains.inVs) {
            const range = this.pendingRanges.get(inV)
            if (range === undefined) {
                this.edgeInserter.do(contains.id, label, contains.outV, inV)
            } else {
                this.pendingRanges.delete(inV)
                this.rangeInserter.do(
                    range.id,
                    contains.outV,
                    range.start.line,
                    range.start.character,
                    range.end.line,
                    range.end.character
                )
            }
        }
    }

    private insertItem(item: item): void {
        for (let inV of item.inVs) {
            if (item.property !== undefined) {
                this.itemInserter.do(item.id, item.outV, inV, item.document, itemPropertyShortForms.get(item.property))
            } else {
                this.itemInserter.do(item.id, item.outV, inV, item.document, null)
            }
        }
    }

    public close(): void {
        this.vertexInserter.finish()
        this.edgeInserter.finish()
        this.rangeInserter.finish()
        this.documentInserter.finish()
        this.itemInserter.finish()
        if (this.pendingRanges.size > 0) {
            console.error(`Pending ranges exists before DB is closed.`)
        }
        this.createIndices()
        this.db.close()
    }
}
