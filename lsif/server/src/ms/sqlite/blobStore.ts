/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as crypto from 'crypto';
import * as fs from 'fs';
import Sqlite from 'better-sqlite3'
import * as uuid from 'uuid';
import * as lsp from 'vscode-languageserver-protocol';
import {
    Edge, Vertex, ElementTypes, VertexLabels, Document, Range, EdgeLabels, contains, Event, EventScope, EventKind, Id, DocumentEvent, FoldingRangeResult,
    RangeBasedDocumentSymbol, DocumentSymbolResult, DiagnosticResult, Moniker, next, ResultSet, moniker, HoverResult, textDocument_hover, textDocument_foldingRange,
    textDocument_documentSymbol, textDocument_diagnostic, MonikerKind, textDocument_declaration, textDocument_definition, textDocument_references, item,
    ItemEdgeProperties, DeclarationResult, DefinitionResult, ReferenceResult, MetaData, nextMoniker, packageInformation, PackageInformation
} from 'lsif-protocol';
import { Compressor, foldingRangeCompressor, CompressorOptions, diagnosticCompressor } from './compress';
import { Inserter } from './inserter';

function assertDefined<T>(value: T | undefined | null): T {
    if (value === undefined || value === null) {
        throw new Error(`Element must be defined`);
    }
    return value;
}

namespace Ranges {
    export function compare(r1: lsp.Range, r2: lsp.Range): number {
        if (r1.start.line < r2.start.line) {
            return -1;
        }
        if (r1.start.line > r2.start.line) {
            return 1;
        }
        if (r1.start.character < r2.start.character) {
            return -1;
        }
        if (r1.start.character > r2.start.character) {
            return 1;
        }
        if (r1.end.line < r2.end.line) {
            return -1;
        }
        if (r1.end.line > r2.end.line) {
            return 1;
        }
        if (r1.end.character < r2.end.character) {
            return -1;
        }
        if (r1.end.character > r2.end.character) {
            return 1;
        }
        return 0;
    }
}

namespace Strings {
    export function compare(s1: string, s2: string): number {
        return ( s1 == s2 ) ? 0 : ( s1 > s2 ) ? 1 : -1;
    }
}

namespace Monikers {
    export function compare(m1: MonikerData, m2: MonikerData): number {
        let result = Strings.compare(m1.identifier, m2.identifier);
        if (result !== 0) {
            return result;
        }
        result = Strings.compare(m1.scheme, m2.scheme);
        if (result !== 0) {
            return result;
        }
        if (m1.kind === m2.kind) {
            return 0;
        }
        const k1 = m1.kind !== undefined ? m1.kind : MonikerKind.import;
        const k2 = m2.kind !== undefined ? m2.kind : MonikerKind.import;
        if (k1 === MonikerKind.import && k2 === MonikerKind.export) {
            return -1;
        }
        if (k1 === MonikerKind.export && k2 === MonikerKind.import) {
            return 1;
        }
        return 0;
    }

    export function isLocal(moniker: MonikerData): boolean {
        return moniker.kind === MonikerKind.local;
    }
}

namespace Diagnostics {
    export function compare(d1: lsp.Diagnostic, d2: lsp.Diagnostic): number {
        let result = Ranges.compare(d1.range, d2.range);
        if (result !== 0) {
            return result;
        }
        result = Strings.compare(d1.message, d2.message);
        if (result !== 0) {
            return result;
        }
        return 0;
    }
}

interface LiteralMap<T> {
    [key: string]: T;
    [key: number]: T;
}

namespace LiteralMap {

    export function create<T = any>(): LiteralMap<T> {
        return Object.create(null);
    }

    export function values<T>(map: LiteralMap<T>): T[] {
        let result: T[] = [];
        for (let key of Object.keys(map)) {
            result.push(map[key]);
        }
        return result;
    }
}

interface RangeData extends Pick<Range, 'start' | 'end' | 'tag'> {
    monikers?: Id[];
    next?: Id;
    hoverResult?: Id;
    declarationResult?: Id;
    definitionResult?: Id;
    referenceResult?: Id;
}

interface ResultSetData {
    monikers?: Id[];
    next?: Id;
    hoverResult?: Id;
    declarationResult?: Id;
    definitionResult?: Id;
    referenceResult?: Id;
}

interface DeclarationResultData {
    values: Id[];
}

interface DefinitionResultData {
    values: Id[];
}

interface ReferenceResultData {
    declarations?: Id[];
    definitions?: Id[];
    references?: Id[];
}

type MonikerData = Pick<Moniker, 'scheme' | 'identifier' | 'kind'> & {
    packageInformation?: Id
}

type PackageInformationData = Pick<
    PackageInformation,
    'name' | 'manager' | 'uri' | 'contents' | 'version' | 'repository'
>

enum BlobFormat {
    utf8Json = 1,
    utf8JsonZipped = 2
}

interface DocumentBlob {
    contents: string
    ranges: LiteralMap<RangeData>
    resultSets?: LiteralMap<ResultSetData>
    monikers?: LiteralMap<MonikerData>
    packageInformation?: LiteralMap<PackageInformationData>
    hovers?: LiteralMap<lsp.Hover>
    declarationResults?: LiteralMap<DeclarationResultData>
    definitionResults?: LiteralMap<DefinitionResultData>
    referenceResults?: LiteralMap<ReferenceResultData>
    foldingRanges?: lsp.FoldingRange[]
    documentSymbols?: lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]
    diagnostics?: lsp.Diagnostic[]
}

interface DataProvider {
    getResultData(id: Id): ResultSetData | undefined;
    removeResultSetData(id: Id): void;
    getMonikerData(id: Id): MonikerData | undefined;
    removeMonikerData(id: Id): void;
    getAndDeleteHoverData(id: Id): lsp.Hover | undefined;
    storeHover(moniker: MonikerData, id: Id): void;
    getAndDeleteDeclarations(declarationResult: Id, documentId: Id): DeclarationResultData | undefined;
    getAndDeleteDefinitions(definitionResult: Id, documentId: Id): DefinitionResultData | undefined;
    getAndDeleteReferences(referencResult: Id, documentId: Id): ReferenceResultData | undefined;
}

type InlineRange = [number, number, number, number];

interface ExternalDefinition {
    scheme: string;
    indentifier: string;
    ranges: InlineRange[];
}

interface ExternalDeclaration {
    scheme: string;
    indentifier: string;
    ranges: InlineRange[];
}

interface ExternalReference {
    scheme: string;
    indentifier: string;
    declarations?: InlineRange[];
    definitions?: InlineRange[];
    references?: InlineRange[];
}

interface DocumentDatabaseData {
    hash: string;
    blob: string;
    declarations?: ExternalDeclaration[];
    definitions?: ExternalDefinition[];
    references?: ExternalReference[];
}

interface MonikerScopedResultData<T> {
    moniker: MonikerData;
    data: T;
}

class DocumentData {

    private provider: DataProvider;

    private id: Id;
    private _uri: string;
    private blob: DocumentBlob;
    private declarations: MonikerScopedResultData<DeclarationResultData>[];
    private definitions: MonikerScopedResultData<DefinitionResultData>[];
    private references: MonikerScopedResultData<ReferenceResultData>[];

    constructor(document: Document, provider: DataProvider) {
        this.provider = provider;
        this.id = document.id;
        this._uri = document.uri;
        this.blob = { contents: Buffer.from(document.contents!, 'base64').toString('utf8'), ranges: Object.create(null) };
        this.declarations = [];
        this.definitions = [];
        this.references = [];
    }

    get uri(): string {
        return this._uri;
    }

    public addRangeData(id: Id, data: RangeData): void {
        this.blob.ranges[id] = data;
        this.addReferencedData(id, data);
    }

    private addResultSetData(id: Id, resultSet: ResultSetData): void {
        if (this.blob.resultSets === undefined) {
            this.blob.resultSets = LiteralMap.create();
        }
        // Many ranges can point to the same result set. Make sure
        // we only travers once.
        if (this.blob.resultSets[id] !== undefined) {
            return;
        }
        this.blob.resultSets[id] = resultSet;
        this.addReferencedData(id, resultSet);
    }

    private addReferencedData(id: Id, item: RangeData | ResultSetData): void {
        const monikers = []

        if (item.monikers !== undefined) {
            for (const itemMoniker of item.monikers) {
                const moniker = assertDefined(this.provider.getMonikerData(itemMoniker));
                this.addMoniker(itemMoniker, moniker);
                monikers.push(moniker)
            }
        }
        if (item.next !== undefined) {
            this.addResultSetData(item.next, assertDefined(this.provider.getResultData(item.next)));
        }

        for (const moniker of monikers) {
            if (item.hoverResult !== undefined) {
                if (Monikers.isLocal(moniker)) {
                    if (this.blob.hovers === undefined) {
                        this.blob.hovers = LiteralMap.create();
                    }
                    if (this.blob.hovers[item.hoverResult] === undefined) {
                        this.blob.hovers[item.hoverResult] = assertDefined(this.provider.getAndDeleteHoverData(item.hoverResult));
                    }
                } else {
                    this.provider.storeHover(moniker, item.hoverResult);
                }
            }
            if (item.declarationResult) {
                const declarations = this.provider.getAndDeleteDeclarations(item.declarationResult, this.id);
                if (declarations !== undefined) {
                    if (Monikers.isLocal(moniker)) {
                        if (this.blob.declarationResults === undefined) {
                            this.blob.declarationResults = LiteralMap.create();
                        }
                        this.blob.declarationResults[item.declarationResult] = declarations;
                    } else {
                        this.declarations.push({ moniker, data: declarations});
                    }
                }
            }
            if (item.definitionResult) {
                const definitions = this.provider.getAndDeleteDefinitions(item.definitionResult, this.id);
                if (definitions !== undefined) {
                    if (Monikers.isLocal(moniker)) {
                        if (this.blob.definitionResults === undefined) {
                            this.blob.definitionResults = LiteralMap.create();
                        }
                        this.blob.definitionResults[item.definitionResult] = definitions;
                    } else {
                        this.definitions.push({ moniker, data: definitions });
                    }
                }
            }
            if (item.referenceResult) {
                const references = this.provider.getAndDeleteReferences(item.referenceResult, this.id);
                if (references !== undefined) {
                    if (Monikers.isLocal(moniker)) {
                        if (this.blob.referenceResults === undefined) {
                            this.blob.referenceResults = LiteralMap.create();
                        }
                        this.blob.referenceResults[item.referenceResult] = references;
                    } else {
                        this.references.push({ moniker, data: references });
                    }
                }
            }
        }
    }

    public addFoldingRangeResult(value: lsp.FoldingRange[]): void {
        this.blob.foldingRanges = value;
    }

    public addDocumentSymbolResult(value: lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]): void {
        this.blob.documentSymbols = value;
    }

    public addDiagnostics(value: lsp.Diagnostic[]): void {
        this.blob.diagnostics = value;
    }

    private addMoniker(id: Id, moniker: MonikerData): void {
        if (this.blob.monikers === undefined) {
            this.blob.monikers = LiteralMap.create();
        }
        this.blob.monikers![id] = moniker
    }

    public addPackageInformation(id: Id, packageInformation: PackageInformationData): void {
        if (this.blob.packageInformation === undefined) {
            this.blob.packageInformation = LiteralMap.create()
        }
        this.blob.packageInformation![id] = packageInformation
    }

    public finalize(): DocumentDatabaseData {
        const id2InlineRange = (id: Id): [number, number, number, number] => {
            const range = this.blob.ranges[id];
            return [range.start.line, range.start.character, range.end.line, range.end.character];
        }

        let externalDeclarations: ExternalDeclaration[] = [];
        for (let declaration of this.declarations) {
            externalDeclarations.push({ scheme: declaration.moniker.scheme, indentifier: declaration.moniker.identifier, ranges: declaration.data.values.map(id2InlineRange) });
        }
        let externalDefinitions: ExternalDefinition[] = [];
        for (let definition of this.definitions) {
            externalDefinitions.push({ scheme: definition.moniker.scheme, indentifier: definition.moniker.identifier, ranges: definition.data.values.map(id2InlineRange) });
        }
        let externalReferences: ExternalReference[] = [];
        for (let reference of this.references) {
            externalReferences.push({
                scheme: reference.moniker.scheme, indentifier: reference.moniker.identifier,
                declarations : reference.data.declarations ? reference.data.declarations.map(id2InlineRange) : undefined,
                definitions : reference.data.definitions ? reference.data.definitions.map(id2InlineRange) : undefined,
                references : reference.data.references ? reference.data.references.map(id2InlineRange) : undefined
            });
        }
        return {
            hash: this.computeHash(),
            blob: JSON.stringify(this.blob, undefined, 0),
            declarations: externalDeclarations.length > 0 ? externalDeclarations : undefined,
            definitions: externalDefinitions.length > 0 ? externalDefinitions : undefined,
            references: externalReferences.length > 0 ? externalReferences : undefined,
        }
    }

    private computeHash(): string {
        const hash = crypto.createHash('md5');
        hash.update(this.blob.contents);
        const options: CompressorOptions = { mode: 'hash' };
        const compressor = assertDefined(Compressor.getVertexCompressor(VertexLabels.range));
        const rangeHashes: Map<Id, string> = new Map();
        for (let key of Object.keys(this.blob.ranges)) {
            const range = this.blob.ranges[key];
            const rangeHash = crypto.createHash('md5').update(JSON.stringify(compressor.compress(range, options), undefined, 0)).digest('base64');
            rangeHashes.set(Number(key), rangeHash);
        }
        for (let item of Array.from(rangeHashes.values()).sort(Strings.compare)) {
            hash.update(item);
        }

        // moniker
        if (this.blob.monikers !== undefined) {
            const monikers = LiteralMap.values(this.blob.monikers).sort(Monikers.compare);
            const compressor = assertDefined(Compressor.getVertexCompressor(VertexLabels.moniker));
            for (let moniker of monikers) {
                const compressed = compressor.compress(moniker, options);
                hash.update(JSON.stringify(compressed, undefined, 0));
            }
        }

        // Assume that folding ranges are already sorted
        if (this.blob.foldingRanges) {
            const compressor = foldingRangeCompressor;
            for (let range of this.blob.foldingRanges) {
                const compressed = compressor.compress(range, options);
                hash.update(JSON.stringify(compressed, undefined, 0))
            }
        }
        // Unsure if we need to sort the children by range or not?
        if (this.blob.documentSymbols && this.blob.documentSymbols.length > 0) {
            const first = this.blob.documentSymbols[0];
            const compressor = lsp.DocumentSymbol.is(first) ? undefined : assertDefined(Compressor.getVertexCompressor(VertexLabels.range));
            if (compressor === undefined) {
                throw new Error(`Document symbol compression not supported`);
            }
            const inline = (result: any[], value: RangeBasedDocumentSymbol) => {
                const item: any[] = [];
                const rangeHash = assertDefined(rangeHashes.get(value.id));
                item.push(rangeHash);
                if (value.children && value.children.length > 0) {
                    const children: any[] = [];
                    for (let child of value.children) {
                        inline(children, child);
                    }
                    item.push(children);
                }
                result.push(item);
            }
            let compressed: any[] = [];
            for (let symbol of (this.blob.documentSymbols as RangeBasedDocumentSymbol[])) {
                inline(compressed, symbol);
            }
            hash.update(JSON.stringify(compressed, undefined, 0));
        }

        // Diagnostics
        if (this.blob.diagnostics && this.blob.diagnostics.length > 0) {
            this.blob.diagnostics = this.blob.diagnostics.sort(Diagnostics.compare);
            const compressor = diagnosticCompressor;
            for (let diagnostic of this.blob.diagnostics) {
                let compressed = compressor.compress(diagnostic, options);
                hash.update(JSON.stringify(compressed, undefined, 0));
            }
        }

        return hash.digest('base64');
    }
}

export class BlobStore implements DataProvider {

    private forceDelete: boolean;
    private version: string;

    private db: Sqlite.Database;
    private documentInserter: Inserter;
    private blobInserter: Inserter;
    private versionInserter: Inserter;
    private declarationInserter: Inserter;
    private definitionInserter: Inserter;
    private referenceInserter: Inserter;
    private hoverInserter: Inserter;

    private knownHashes: Set<string>;
    private insertedBlobs: Set<string>;
    private insertedHovers: Set<string>;

    private documents: Map<Id, Document>;
    private documentDatas: Map<Id, DocumentData | null>;

    private foldingRanges: Map<Id, lsp.FoldingRange[]>;
    private documentSymbols: Map<Id, lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]>;
    private diagnostics: Map<Id, lsp.Diagnostic[]>;

    private rangeDatas: Map<Id, RangeData>;
    private resultSetDatas: Map<Id, ResultSetData>;
    private monikerDatas: Map<Id, MonikerData>;
    private monikerSets: Map<Id, Set<Id>>;
    private monikerAttachments: Map<Id, Id>;
    private packageInformationDatas: Map<Id, PackageInformationData>;
    private hoverDatas: Map<Id, lsp.Hover>;
    private declarationDatas: Map<Id /* result id */, Map<Id /* document id */, DeclarationResultData>>;
    private definitionDatas: Map<Id /* result id */, Map<Id /* document id */, DefinitionResultData>>;
    private referenceDatas: Map<Id /* result id */, Map<Id /* document id */, ReferenceResultData>>;

    private containsDatas: Map<Id, Id[]>;

    constructor(filename: string, version: string, forceDelete: boolean = false) {
        this.forceDelete = forceDelete;
        this.version = version;
        this.knownHashes = new Set();
        this.insertedBlobs = new Set();
        this.insertedHovers = new Set();

        this.documents = new Map();
        this.documentDatas = new Map();

        this.foldingRanges = new Map();
        this.documentSymbols = new Map();
        this.diagnostics = new Map();

        this.rangeDatas = new Map();
        this.resultSetDatas = new Map();
        this.monikerDatas = new Map();
        this.monikerSets = new Map();
        this.monikerAttachments = new Map();
        this.packageInformationDatas = new Map();
        this.hoverDatas = new Map();
        this.declarationDatas = new Map();
        this.definitionDatas = new Map();
        this.referenceDatas = new Map();
        this.containsDatas = new Map();

        if (forceDelete) {
            try {
                fs.unlinkSync(filename);
            } catch (err) {
            }
        } else {
            forceDelete = !fs.existsSync(filename);
        }
        this.db = new Sqlite(filename);
        this.db.pragma('synchronous = OFF');
        this.db.pragma('journal_mode = MEMORY');
        if (forceDelete) {
            this.createTables();
        }
        this.db.prepare(`Insert Into versionTags (tag, dateTime) Values (?, ?)`).run(this.version, Date.now());
        this.blobInserter = new Inserter(this.db, 'Insert Into blobs (hash, format, content)', 3, 16);
        this.documentInserter = new Inserter(this.db, 'Insert Into documents (documentHash, uri)', 2, 16);
        this.versionInserter = new Inserter(this.db, 'Insert Into versions (version, hash)', 2, 256);
        this.declarationInserter = new Inserter(this.db, 'Insert Into decls (scheme, identifier, documentHash, startLine, startCharacter, endLine, endCharacter)', 7, 128);
        this.definitionInserter = new Inserter(this.db, 'Insert Into defs (scheme, identifier, documentHash, startLine, startCharacter, endLine, endCharacter)', 7, 128);
        this.referenceInserter = new Inserter(this.db, 'Insert Into refs (scheme, identifier, documentHash, kind, startLine, startCharacter, endLine, endCharacter)', 8, 64);
        this.hoverInserter = new Inserter(this.db, 'Insert Into hovers (scheme, identifier, hoverHash)', 3, 128);

        if (!forceDelete) {
            const hashes: { hash: string }[] = this.db.prepare('Select hash From blobs').all();
            for (let item of hashes) {
                this.knownHashes.add(item.hash);
            }
        }
    }

    private createTables(): void {
        this.db.exec('Create Table meta (id Integer Unique Primary Key, value Text Not Null)');
        this.db.exec('Create Table format (format Text Not Null)');
        this.db.exec('Create Table versionTags (tag Text Unique Primary Key, dateTime Integer Not Null)');
        this.db.exec('Create Table blobs (hash Text Unique Primary Key, format Integer Not Null, content Blob Not Null)');
        this.db.exec('Create Table documents (documentHash Text Not Null, uri Text Not Null)');
        this.db.exec('Create Table versions (version Text Not Null, hash Text Not Null)');
        this.db.exec('Create Table decls (scheme Text Not Null, identifier Text Not Null, documentHash Text Not Null, startLine Integer Not Null, startCharacter Integer Not Null, endLine Integer Not Null, endCharacter Integer Not Null)');
        this.db.exec('Create Table defs (scheme Text Not Null, identifier Text Not Null, documentHash Text Not Null, startLine Integer Not Null, startCharacter Integer Not Null, endLine Integer Not Null, endCharacter Integer Not Null)');
        this.db.exec('Create Table refs (scheme Text Not Null, identifier Text Not Null, documentHash Text Not Null, kind Integer Not Null, startLine Integer Not Null, startCharacter Integer Not Null, endLine Integer Not Null, endCharacter Integer Not Null)');
        this.db.exec('Create Table hovers (scheme Text Not Null, identifier Text Not Null, hoverHash Text Not Null)');
    }

    private createIndices(): void {
        this.db.exec('Create Index _versionTags_tag on versionTags (tag)');
        this.db.exec('Create Index _versionTags_dateTime on versionTags (dateTime)');
        this.db.exec('Create Index _blobs on blobs (hash)');
        this.db.exec('Create Index _documents on documents (documentHash)');
        this.db.exec('Create Index _versions on versions (version)');
        this.db.exec('Create Index _decls on decls (identifier, scheme, documentHash)');
        this.db.exec('Create Index _defs on defs (identifier, scheme, documentHash)');
        this.db.exec('Create Index _refs on refs (identifier, scheme, documentHash)');
        this.db.exec('Create Index _hovers on hovers (identifier, scheme, hoverHash)');
    }

    public insert(element: Edge | Vertex): void {
        if (element.type === ElementTypes.vertex) {
            switch(element.label) {
                case VertexLabels.metaData:
                    this.handleMetaData(element);
                    break;
                case VertexLabels.document:
                    this.documents.set(element.id, element);
                    break;
                case VertexLabels.range:
                    this.handleRange(element);
                    break;
                case VertexLabels.resultSet:
                    this.handleResultSet(element);
                    break;
                case VertexLabels.moniker:
                    this.handleMoniker(element);
                    break;
                case VertexLabels.packageInformation:
                    this.handlePackageInformation(element)
                    break;
                case VertexLabels.hoverResult:
                    this.handleHover(element);
                    break;
                case VertexLabels.declarationResult:
                    this.handleDeclarationResult(element);
                    break;
                case VertexLabels.definitionResult:
                    this.handleDefinitionResult(element);
                    break;
                case VertexLabels.referenceResult:
                    this.handleReferenceResult(element);
                    break;
                case VertexLabels.foldingRangeResult:
                    this.handleFoldingRange(element);
                    break;
                case VertexLabels.documentSymbolResult:
                    this.handleDocumentSymbols(element);
                    break;
                case VertexLabels.diagnosticResult:
                    this.handleDiagnostics(element);
                    break;
                case VertexLabels.event:
                    this.handleEvent(element);
                    break;
                case VertexLabels.packageInformation:
                    console.log('packageInformation vertex unimplemented')
                    break
            }
        } else if (element.type === ElementTypes.edge) {
            switch(element.label) {
                case EdgeLabels.next:
                    this.handleNextEdge(element);
                    break;
                case EdgeLabels.moniker:
                    this.handleMonikerEdge(element)
                    break;
                case EdgeLabels.nextMoniker:
                    this.handleNextMonikerEdge(element)
                    break;
                case EdgeLabels.packageInformation:
                    this.handlePackageInformationEdge(element)
                    break;
                case EdgeLabels.textDocument_foldingRange:
                    this.handleFoldingRangeEdge(element);
                    break;
                case EdgeLabels.textDocument_documentSymbol:
                    this.handleDocumentSymbolEdge(element);
                    break;
                case EdgeLabels.textDocument_diagnostic:
                    this.handleDiagnosticsEdge(element);
                    break;
                case EdgeLabels.textDocument_hover:
                    this.handleHoverEdge(element);
                    break;
                case EdgeLabels.textDocument_declaration:
                    this.handleDeclarationEdge(element);
                    break;
                case EdgeLabels.textDocument_definition:
                    this.handleDefinitionEdge(element);
                    break;
                case EdgeLabels.textDocument_references:
                    this.handleReferenceEdge(element);
                    break;
                case EdgeLabels.item:
                    this.handleItemEdge(element);
                    break;
                case EdgeLabels.contains:
                    this.handleContains(element);
                    break;
            }
        }
    }

    public getResultData(id: Id): ResultSetData | undefined {
        return this.resultSetDatas.get(id);
    }

    public removeResultSetData(id: Id): void {
        this.resultSetDatas.delete(id);
    }

    public getMonikerData(id: Id): MonikerData | undefined {
        return this.monikerDatas.get(id);
    }

    public removeMonikerData(id: Id): void {
        this.monikerDatas.delete(id);
    }

    public getAndDeleteHoverData(id: Id): lsp.Hover | undefined {
        let result = this.hoverDatas.get(id);
        if (result !== undefined) {
            // We don't delete the hover right now.
            // See https://github.com/microsoft/lsif-node/issues/57
            // this.hoverDatas.delete(id);
        }
        return result;
    }

    public getAndDeleteDeclarations(declaratinResult: Id, documentId: Id): DeclarationResultData | undefined {
        const map = assertDefined(this.declarationDatas.get(declaratinResult));
        const result = map.get(documentId);
        map.delete(documentId);
        return result;
    }

    public getAndDeleteDefinitions(definitionResult: Id, documentId: Id): DefinitionResultData | undefined {
        const map = assertDefined(this.definitionDatas.get(definitionResult));
        const result = map.get(documentId);
        map.delete(documentId);
        return result;
    }

    public getAndDeleteReferences(referenceResult: Id, documentId: Id): ReferenceResultData | undefined {
        const map = assertDefined(this.referenceDatas.get(referenceResult));
        const result = map.get(documentId);
        map.delete(documentId);
        return result;
    }

    public runInsertTransaction(cb: (db: BlobStore) => void): void {
        cb(this);
    }

    public close(): void {
        this.blobInserter.finish();
        this.documentInserter.finish();
        this.versionInserter.finish();
        this.declarationInserter.finish();
        this.definitionInserter.finish();
        this.referenceInserter.finish();
        this.hoverInserter.finish();
        if (this.forceDelete) {
            this.createIndices();
        }
        this.db.close();
    }

    private handleMetaData(vertex: MetaData): void {
        if (!this.forceDelete) {
            return;
        }
        let value = JSON.stringify(vertex, undefined, 0);
        this.db.exec(`Insert Into meta (id, value) Values (${vertex.id}, '${value}')`);
        this.db.exec(`Insert into format (format) Values ('blob')`);
    }

    private handleEvent(event: Event): void {
        if (event.scope === EventScope.project) {

        } else if (event.scope === EventScope.document) {
            let documentEvent = event as DocumentEvent;
            switch (event.kind) {
                case EventKind.begin:
                    this.handleDocumentBegin(documentEvent);
                    break;
                case EventKind.end:
                    this.handleDocumentEnd(documentEvent);
                    break;
            }
        }
    }

    private handleDocumentBegin(event: DocumentEvent) {
        const document = this.documents.get(event.data);
        if (document === undefined) {
            throw new Error(`Document with id ${event.data} not known`);
        }
        this.getOrCreateDocumentData(document);
        this.documents.delete(event.data);
    }

    private handleRange(range: Range): void {
        let data: RangeData = { start: range.start, end: range.end, tag: range.tag };
        this.rangeDatas.set(range.id, data);
    }

    private handleResultSet(set: ResultSet): void {
        let data: ResultSetData = {};
        this.resultSetDatas.set(set.id, data);
    }

    private handleMoniker(moniker: Moniker): void {
        let data: MonikerData = { scheme: moniker.scheme, identifier: moniker.identifier, kind: moniker.kind };
        this.monikerDatas.set(moniker.id, data);
    }

    private handleMonikerEdge(moniker: moniker): void {
        assertDefined(this.rangeDatas.get(moniker.outV) || this.resultSetDatas.get(moniker.outV));
        assertDefined(this.monikerDatas.get(moniker.inV))
        this.monikerAttachments.set(moniker.outV, moniker.inV)
        this.updateMonikerSets([moniker.inV])
    }

    private handleNextMonikerEdge(nextMoniker: nextMoniker): void {
        assertDefined(this.monikerDatas.get(nextMoniker.inV))
        assertDefined(this.monikerDatas.get(nextMoniker.outV))
        this.updateMonikerSets([nextMoniker.inV, nextMoniker.outV])
    }

    private updateMonikerSets(vals: Id[]): void {
        const combined = new Set<Id>()
        for (const val of vals) {
            combined.add(val)

            for (const v of this.monikerSets.get(val) || new Set()) {
                combined.add(v)
            }
        }

        for (const val of combined) {
            this.monikerSets.set(val, combined)
        }
    }

    private handlePackageInformation(packageInformation: PackageInformation): void {
        let data: PackageInformationData = {
            name: packageInformation.name,
            manager: packageInformation.manager,
            uri: packageInformation.uri,
            contents: packageInformation.contents,
            version: packageInformation.version,
            repository: packageInformation.repository,
        }

        this.packageInformationDatas.set(packageInformation.id, data)
    }

    private handlePackageInformationEdge(packageInformation: packageInformation): void {
        const source: MonikerData = assertDefined(this.monikerDatas.get(packageInformation.outV))
        assertDefined(this.packageInformationDatas.get(packageInformation.inV))
        source.packageInformation = packageInformation.inV
    }

    private handleHover(hover: HoverResult): void {
        this.hoverDatas.set(hover.id, hover.result);
    }

    private handleHoverEdge(edge: textDocument_hover): void {
        const outV: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV));
        this.ensureMoniker(outV);
        assertDefined(this.hoverDatas.get(edge.inV));
        outV.hoverResult = edge.inV;
    }

    private handleDeclarationResult(result: DeclarationResult): void {
        this.declarationDatas.set(result.id, new Map());
    }

    private handleDeclarationEdge(edge: textDocument_declaration): void {
        const outV: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV));
        this.ensureMoniker(outV);
        assertDefined(this.declarationDatas.get(edge.inV));
        outV.declarationResult = edge.inV;
    }

    private handleDefinitionResult(result: DefinitionResult): void {
        this.definitionDatas.set(result.id, new Map());
    }

    private handleDefinitionEdge(edge: textDocument_definition): void {
        const outV: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV));
        this.ensureMoniker(outV);
        assertDefined(this.definitionDatas.get(edge.inV));
        outV.definitionResult = edge.inV;
    }

    private handleReferenceResult(result: ReferenceResult): void {
        this.referenceDatas.set(result.id, new Map());
    }

    private handleReferenceEdge(edge: textDocument_references): void {
        const outV: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV));
        this.ensureMoniker(outV);
        assertDefined(this.referenceDatas.get(edge.inV));
        outV.referenceResult = edge.inV;
    }

    private ensureMoniker(data: RangeData | ResultSetData): void {
        if (data.monikers !== undefined) {
            return;
        }
        const monikerData: MonikerData = { scheme: '$synthetic', identifier: uuid.v4() };
        data.monikers = [monikerData.identifier];
        this.monikerDatas.set(monikerData.identifier, monikerData);
    }

    private handleItemEdge(edge: item): void {
        let property: ItemEdgeProperties | undefined = edge.property;
        if (property === undefined) {
            const map: Map<Id, DefinitionResultData> | Map<Id, DeclarationResultData> = assertDefined(this.declarationDatas.get(edge.outV) || this.definitionDatas.get(edge.outV));
            let data: DefinitionResultData | DeclarationResultData | undefined = map.get(edge.document);
            if (data === undefined) {
                data = { values: edge.inVs.slice() };
                map.set(edge.document, data);
            } else {
                data.values.push(...edge.inVs);
            }
        } else {
            const map: Map<Id, ReferenceResultData> = assertDefined(this.referenceDatas.get(edge.outV));
            let data: ReferenceResultData | undefined = map.get(edge.document);
            if (data === undefined) {
                data = {};
                map.set(edge.document, data);
            }
            switch (property) {
                case ItemEdgeProperties.declarations:
                    if (data.declarations === undefined) {
                        data.declarations = edge.inVs.slice();
                    } else {
                        data.declarations.push(...edge.inVs);
                    }
                    break;
                case ItemEdgeProperties.definitions:
                    if (data.definitions === undefined) {
                        data.definitions = edge.inVs.slice();
                    } else {
                        data.definitions.push(...edge.inVs);
                    }
                    break;
                case ItemEdgeProperties.references:
                    if (data.references === undefined) {
                        data.references = edge.inVs.slice();
                    } else {
                        data.references.push(...edge.inVs);
                    }
                    break;
            }
        }
    }

    private handleFoldingRange(folding: FoldingRangeResult): void {
        this.foldingRanges.set(folding.id, folding.result);
    }

    private handleFoldingRangeEdge(edge: textDocument_foldingRange): void {
        const source = assertDefined(this.getDocumentData(edge.outV));
        source.addFoldingRangeResult(assertDefined(this.foldingRanges.get(edge.inV)));
    }

    private handleDocumentSymbols(symbols: DocumentSymbolResult): void {
        this.documentSymbols.set(symbols.id, symbols.result);
    }

    private handleDocumentSymbolEdge(edge: textDocument_documentSymbol): void {
        const source = assertDefined(this.getDocumentData(edge.outV));
        source.addDocumentSymbolResult(assertDefined(this.documentSymbols.get(edge.inV)));
    }

    private handleDiagnostics(diagnostics: DiagnosticResult): void {
        this.diagnostics.set(diagnostics.id, diagnostics.result);
    }

    private handleDiagnosticsEdge(edge: textDocument_diagnostic): void {
        const source = assertDefined(this.getDocumentData(edge.outV));
        source.addDiagnostics(assertDefined(this.diagnostics.get(edge.inV)));
    }

    private handleNextEdge(edge: next): void {
        const outV: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV));
        assertDefined(this.resultSetDatas.get(edge.inV));
        outV.next = edge.inV;
    }

    private handleContains(contains: contains): boolean {
        let values = this.containsDatas.get(contains.outV);
        if (values === undefined) {
            values = [];
            this.containsDatas.set(contains.outV, values);
        }
        values.push(...contains.inVs);
        return true;
    }

    private handleDocumentEnd(event: DocumentEvent) {
        for (const [key, value] of this.monikerAttachments.entries()) {
            const ids = this.monikerSets.get(value)
            if (!ids) {
                throw new Error("moniker set is empty")
            }

            const source: RangeData | ResultSetData = assertDefined(this.rangeDatas.get(key) || this.resultSetDatas.get(key));
            ids.forEach(id => assertDefined(this.monikerDatas.get(id)))
            source.monikers = Array.from(ids)
        }

        const documentData = this.getEnsureDocumentData(event.data);
        const contains = this.containsDatas.get(event.data);
        if (contains !== undefined) {
            for (let id of contains) {
                const range = assertDefined(this.rangeDatas.get(id));
                documentData.addRangeData(id, range);
            }
        }
        for (const [key, value] of this.packageInformationDatas) {
            documentData.addPackageInformation(key, value)
        }
        let data = documentData.finalize()
        if (this.knownHashes.has(data.hash)) {
            this.versionInserter.do(this.version, data.hash);
        } else {
            this.blobInserter.do(data.hash, BlobFormat.utf8Json,  Buffer.from(data.blob, 'utf8'));
            this.documentInserter.do(data.hash, documentData.uri);
            this.versionInserter.do(this.version, data.hash);
            if (data.declarations) {
                for (let declaration of data.declarations) {
                    for (let range in declaration.ranges) {
                        this.declarationInserter.do(declaration.scheme, declaration.indentifier, data.hash, range[0], range[1], range[2], range[3]);
                    }
                }
            }
            if (data.definitions) {
                for (let definition of data.definitions) {
                    for (let range of definition.ranges) {
                        this.definitionInserter.do(definition.scheme, definition.indentifier, data.hash, range[0], range[1], range[2], range[3]);
                    }
                }
            }
            if (data.references) {
                for (let reference of data.references) {
                    if (reference.declarations) {
                        for (let range of reference.declarations) {
                            this.referenceInserter.do(reference.scheme, reference.indentifier, data.hash, 0, range[0], range[1], range[2], range[3]);
                        }
                    }
                    if (reference.definitions) {
                        for (let range of reference.definitions) {
                            this.referenceInserter.do(reference.scheme, reference.indentifier, data.hash, 1, range[0], range[1], range[2], range[3]);
                        }
                    }
                    if (reference.references) {
                        for (let range of reference.references) {
                            this.referenceInserter.do(reference.scheme, reference.indentifier, data.hash, 2, range[0], range[1], range[2], range[3]);
                        }
                    }
                }
            }
        }
        this.documentDatas.set(event.id, null);
    }

    public storeHover(moniker: MonikerData, id: Id): void {
        let hover = this.getAndDeleteHoverData(id);
        // We have already processed the hover
        if (hover === undefined) {
            return;
        }
        const blob = JSON.stringify(hover, undefined, 0);
        const blobHash = crypto.createHash('md5').update(blob).digest('base64');
        // Actuall
        if (this.knownHashes.has(blobHash)) {
            this.versionInserter.do(this.version, blobHash);
        } else {
            if (!this.insertedBlobs.has(blobHash)) {
                this.blobInserter.do(blobHash, BlobFormat.utf8Json, Buffer.from(blob, 'utf8'));
                this.insertedBlobs.add(blobHash);
            }
            const hoverHash = crypto.createHash('md5').update(
                JSON.stringify({s: moniker.scheme, i: moniker.identifier, b: blobHash }, undefined, 0))
            .digest('base64');
            if (!this.insertedHovers.has(hoverHash)) {
                this.hoverInserter.do(moniker.scheme, moniker.identifier, blobHash);
                this.versionInserter.do(this.version, blobHash);
                this.insertedHovers.add(hoverHash);
            }
        }
    }

    private getOrCreateDocumentData(document: Document): DocumentData {
        let result: DocumentData | undefined | null = this.documentDatas.get(document.id);
        if (result === null) {
            throw new Error(`The document ${document.uri} has already been processed`);
        }
        result = new DocumentData(document, this);
        this.documentDatas.set(document.id, result);
        return result;
    }

    private getDocumentData(id: Id): DocumentData | undefined {
        let result: DocumentData | undefined | null = this.documentDatas.get(id);
        if (result === null) {
            throw new Error(`The document with Id ${id} has already been processed.`);
        }
        return result;
    }

    private getEnsureDocumentData(id: Id): DocumentData {
        let result: DocumentData | undefined | null = this.documentDatas.get(id);
        if (result === undefined) {
            throw new Error(`No document data found for id ${id}`);
        }
        if (result === null) {
            throw new Error(`The document with Id ${id} has already been processed.`);
        }
        return result;
    }
}
