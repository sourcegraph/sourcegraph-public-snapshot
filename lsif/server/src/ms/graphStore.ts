/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as lsp from 'vscode-languageserver';
import Sqlite from 'better-sqlite3';
import { CompressionKind, CompressorDescription, MetaData } from './protocol.compress';
import { Database, UriTransformer } from './database';
import {
    DeclarationResult,
    DefinitionResult,
    DocumentSymbolResult,
    EdgeLabels,
    FoldingRangeResult,
    HoverResult,
    Id,
    ItemEdgeProperties,
    Range,
    RangeBasedDocumentSymbol,
    ReferenceResult
    } from 'lsif-protocol';
import { DocumentInfo } from './files';
import { URI } from 'vscode-uri';

interface DecompressorPropertyDescription {
	name: string;
	index: number;
	compressionKind: CompressionKind
	longForm?: Map<string | number, string>;
}

class Decompressor {

	public static all: Map<number, Decompressor> = new Map();

	public static get(id: number): Decompressor | undefined {
		return this.all.get(id);
	}

	private id: number;
	private parentId: number | undefined;
	private parent: Decompressor | undefined;
	private properties: DecompressorPropertyDescription[];

	constructor(description: CompressorDescription) {
		this.id = description.id;
		this.parentId = description.parent;
		this.properties = [];
		for (let item of description.properties) {
			let propertyDescription: DecompressorPropertyDescription = {
				name: item.name,
				index: item.index,
				compressionKind: item.compressionKind,
				longForm: undefined
			};
			if (item.shortForm !== undefined) {
				propertyDescription.longForm = new Map();
				for (let element of item.shortForm) {
					propertyDescription.longForm.set(element[1], element[0]);
				}
			}
			this.properties.push(propertyDescription);
		}
		Decompressor.all.set(this.id, this);
	}

	public link(): void {
		if (this.parentId !== undefined) {
			this.parent = Decompressor.get(this.parentId);
		}
	}

	public getPropertyDescription(name: string): DecompressorPropertyDescription | undefined {
		for (let item of this.properties) {
			if (item.name === name) {
				return item;
			}
		}
		return undefined;
	}

	public decompress<T = object>(compressed: any[]): T {
		let result = this.parent !== undefined ? this.parent.decompress(compressed) : Object.create(null);
		for (let property of this.properties) {
			let index = property.index;
			let value = compressed[index];
			if (value === null || value === undefined) {
				continue;
			}
			let decompressor: Decompressor | undefined;
			switch (property.compressionKind) {
				case CompressionKind.raw:
					result[property.name] = value;
					break;
				case CompressionKind.scalar:
					let convertedScalar = value;
					if (property.longForm !== undefined) {
						let long = property.longForm.get(value);
						if (long !== undefined) {
							convertedScalar = long;
						}
					}
					let dotIndex = property.name.indexOf('.');
					if (dotIndex !== -1) {
						let container = property.name.substr(0, dotIndex);
						let name = property.name.substring(dotIndex + 1);
						if (result[container] === undefined) {
							result[container] = Object.create(null);
						}
						result[container][name] = convertedScalar;
					} else {
						result[property.name] = convertedScalar;
					}
					break;
				case CompressionKind.literal:
					if (!Array.isArray(value) || typeof value[0] !== 'number') {
						throw new Error(`Compression kind literal detected on non array value. The property is ${property.name}`);
					}
					let convertedLiteral: any;
					decompressor = Decompressor.get(value[0]);
					if (decompressor === undefined) {
						throw new Error(`No decompression found for property ${property.name} and id ${value[0]}`);
					}
					convertedLiteral = decompressor.decompress(value);
					result[property.name] = convertedLiteral;
					break;
				case CompressionKind.array:
					if (!Array.isArray(value)) {
						throw new Error(`Compression kind array detected on non array value. The property is ${property.name}`);
					}
					let convertedArray: any[] = [];
					for (let element of value) {
						let type = typeof element;
						if (type === 'string' || type === 'number' || type === 'boolean') {
							convertedArray.push(element);
						} else if (Array.isArray(element) && element.length > 0 && typeof element[0] === 'number') {
							decompressor = Decompressor.get(element[0]);
							if (decompressor === undefined) {
								throw new Error(`No decompression found for property ${property.name} and id ${element[0]}`)
							}
							convertedArray.push(decompressor.decompress(element));
						} else {
							throw new Error(`The array element is neither a scalar nor an array.`);
						}
					}
					result[property.name] = convertedArray;
					break;
				case CompressionKind.any:
					let convertedAny: any;
					let type = typeof value;
					if (type === 'string' || type === 'number' || type === 'boolean') {
						convertedAny = value;
					} else if (Array.isArray(value)) {
						convertedAny = [];
						for (let element of value) {
							let type = typeof element;
							if (type === 'string' || type === 'number' || type === 'boolean') {
								(convertedAny as any[]).push(element);
							} else if (Array.isArray(element) && element.length > 0 && typeof element[0] === 'number') {
								decompressor = Decompressor.get(element[0]);
								if (decompressor === undefined) {
									throw new Error(`No decompression found for property ${property.name} and id ${element[0]}`)
								}
								(convertedAny as any[]).push(decompressor.decompress(element));
							} else {
								throw new Error(`The array element is neither a scalar nor an array.`);
							}
						}
					}
					if (convertedAny === undefined) {
						throw new Error(`Comression kind any can't be handled for property ${property.name}. Value is ${JSON.stringify(value)}`);
					}
					result[property.name] = convertedAny;
					break;
				default:
					throw new Error(`Compression kind ${property.compressionKind} unknown.`)
			}
		}
		return result;
	}
}

interface RangeResult {
	id: number;
	startLine: number;
	startCharacter: number;
	endLine: number;
	endCharacter: number;
}

interface MetaDataResult {
	id: number;
	value: string;
}

interface IdResult {
	id: Id;
}

interface LocationResult extends IdResult {
	uri: string;
	startLine: number;
	startCharacter: number;
	endLine: number;
	endCharacter: number;
}

interface LocationResultWithProperty extends LocationResult {
	property: number;
}

interface DocumentResult extends IdResult {
	label: number;
	value: string;
}

interface VertexResult extends IdResult {
	label: number;
	value: string;
}

interface ContentResult extends IdResult {
	content: string;
}

abstract class Retriever<T extends IdResult> {

	private values: Id[];

	public constructor(private name: string, private db: Sqlite.Database, private batchSize: number) {
		this.values= [];
	}

	public clear(): void {
		this.values = [];
	}

	public get isEmpty(): boolean {
		return this.values.length === 0;
	}

	public add(id: Id): void {
		this.values.push(id);
	}

	public addMany(ids: Id[]): void {
		this.values.push(...ids);
	}

	public run(): T[] {
		let result: T[] = new Array(this.values.length);
		let batch: Id[] = [];
		let mapping: Map<Id, number> = new Map();
		for (let i = 0; i < this.values.length; i++) {
			let value = this.values[i];
			batch.push(value);
			mapping.set(value, i);
			if (batch.length === this.batchSize) {
				this.retrieveBatch(result, batch, mapping);
				batch = [];
				mapping.clear();
			}
		}
		if (batch.length > 0) {
			this.retrieveBatch(result, batch, mapping);
		}
		this.values = [];
		return result;
	}

	private retrieveBatch(result: T[], batch: Id[], mapping: Map<Id, number>): void {
		let stmt = batch.length === this.batchSize
			? this.getFullStatement(this.batchSize)
			: this.getRestStatement(batch.length);

		let data: T[] = stmt.all(batch);
		if (batch.length !== data.length) {
			throw new Error(`Couldn't retrieve all data for retriever ${this.name}`);
		}
		for (let element of data) {
			result[mapping.get(element.id)!] = element;
		}
	}

	protected prepare(stmt: string, size: number): Sqlite.Statement {
		return this.db.prepare(`${stmt} (${new Array(size).fill('?').join(',')})`)
	}

	protected abstract getFullStatement(size: number): Sqlite.Statement;

	protected abstract getRestStatement(size: number): Sqlite.Statement;
}

class VertexRetriever extends Retriever<VertexResult> {

	private static statement: string = [
		'Select v.id, v.label, v.value from vertices v',
		'Where v.id in'
	].join(' ');

	private static preparedStatements: Map<number, Sqlite.Statement> = new Map();

	public constructor(db: Sqlite.Database, batchSize: number = 16) {
		super('VertexRetriever', db, batchSize)
	}

	protected getFullStatement(size: number): Sqlite.Statement {
		let result = VertexRetriever.preparedStatements.get(size);
		if (!result) {
			result = this.prepare(VertexRetriever.statement, size);
			VertexRetriever.preparedStatements.set(size, result);
		}
		return result;
	}

	protected getRestStatement(size: number): Sqlite.Statement {
		return this.prepare(VertexRetriever.statement, size);
	}
}

export class GraphStore extends Database {

	private db!: Sqlite.Database;

	private allDocumentsStmt!: Sqlite.Statement;
	private getDocumentContentStmt!: Sqlite.Statement;
	private findRangeStmt!: Sqlite.Statement;
	private findDocumentStmt!: Sqlite.Statement;
	private findResultStmt!: Sqlite.Statement;
	private findResultViaSetStmt!: Sqlite.Statement;
	private findResultForDocumentStmt!: Sqlite.Statement;
	private findRangeFromReferenceResult!: Sqlite.Statement;
	private findResultFromReferenceResult!: Sqlite.Statement;
	private findRangeFromResult!: Sqlite.Statement;

	private projectRoot!: URI;
	private vertexLabels: Map<string, number> | undefined;
	private edgeLabels: Map<string, number> | undefined;
	private itemEdgeProperties: Map<string, number> | undefined;

	public constructor() {
		super();
	}

	public load(file: string, transformerFactory: (projectRoot: string) => UriTransformer): Promise<void> {
		this.db = new Sqlite(file, { readonly: true });
		this.readMetaData();
		this.allDocumentsStmt = this.db.prepare('Select id, uri From documents');
		this.getDocumentContentStmt = this.db.prepare('Select content From contents Where id = ?');
		this.findDocumentStmt = this.db.prepare('Select id From documents Where uri = ?');
		this.findRangeStmt = this.db.prepare([
			'Select r.id, r.startLine, r.startCharacter, r.endline, r.endCharacter From ranges r',
			'Inner Join documents d On r.belongsTo = d.id',
			'where',
				'd.uri = $uri and (',
					'(r.startLine < $line and $line < r.endline) or',
					'(r.startLine = $line and r.startCharacter <= $character and $line < r.endline) or',
					'(r.startLine < $line and r.endLine = $line and $character <= r.endCharacter) or',
					'(r.startLine = $line and r.endLine = $line and r.startCharacter <= $character and $character <= r.endCharacter)',
			  	')'
		].join(' '));
		let nextLabel = this.edgeLabels !== undefined ? this.edgeLabels.get(EdgeLabels.next)! : EdgeLabels.next
		this.findResultStmt = this.db.prepare([
			'Select v.id, v.label, v.value From vertices v',
			'Inner join edges e On e.inV = v.id',
			'Where e.outV = $source and e.label = $label'
		].join(' '));
		this.findResultViaSetStmt = this.db.prepare([
			'Select v.id, v.label, v.value from edges e1',
			'Inner Join edges e2 On e1.inv = e2.outV',
			'Inner Join vertices v On e2.inV = v.id',
			`where e1.outV = $source and e1.label = ${nextLabel} and e2.label = $label`

		].join(' '));
		this.findResultForDocumentStmt = this.db.prepare([
			'Select v.id, v.label, v.value from vertices v',
			'Inner Join edges e On e.inV = v.id',
			'Inner Join documents d On d.id = e.outV',
			'Where d.uri = $uri and e.label = $label'
		].join(' '));

		this.findRangeFromResult = this.db.prepare([
			'Select r.id, r.startLine, r.startCharacter, r.endLine, r.endCharacter, d.uri from ranges r',
			'Inner Join items i On i.inV = r.id',
			'Inner Join documents d On r.belongsTo = d.id',
			'Where i.outV = $id'
		].join(' '));
		this.findRangeFromReferenceResult = this.db.prepare([
			'Select r.id, r.startLine, r.startCharacter, r.endLine, r.endCharacter, i.property, d.uri from ranges r',
			'Inner Join items i On i.inV = r.id',
			'Inner Join documents d On r.belongsTo = d.id',
			'Where i.outV = $id and (i.property in (1, 2, 3))'
		].join(' '));
		this.findResultFromReferenceResult = this.db.prepare([
			'Select v.id, v.label, v.value from vertices v',
			'Inner Join items i On i.inV = v.id',
			'Where i.outV = $id and i.property = 4'
		].join(' '));
		this.initialize(transformerFactory);
		return Promise.resolve();
	}

	private readMetaData(): void {
		let result: MetaDataResult[] = this.db.prepare('Select * from meta').all();
		if (result === undefined || result.length !== 1) {
			throw new Error('Failed to read meta data record.');
		}
		let metaData: MetaData = JSON.parse(result[0].value);
		if (metaData.projectRoot === undefined) {
			throw new Error('No project root provided.');
		}
		this.projectRoot = URI.parse(metaData.projectRoot)
		if (metaData.compressors !== undefined) {
			this.vertexLabels = new Map();
			this.edgeLabels = new Map();
			this.itemEdgeProperties = new Map();
			for (let decription of metaData.compressors.all) {
				new Decompressor(decription);
			}
			for (let element of Decompressor.all.values()) {
				element.link();
			}
			// Vertex Compressor
			let decompressor = Decompressor.get(metaData.compressors.vertexCompressor);
			if (decompressor === undefined) {
				throw new Error('No vertex decompressor found.');
			}
			let description = decompressor.getPropertyDescription('label');
			if (description === undefined || description.longForm === undefined) {
				throw new Error('No vertex label property description found.');
			}
			for (let item of description.longForm) {
				this.vertexLabels.set(item[1], item[0] as number);
			}
			// Edge Compressor
			decompressor = Decompressor.get(metaData.compressors.edgeCompressor);
			if (decompressor === undefined) {
				throw new Error('No edge decompressor found.');
			}
			description = decompressor.getPropertyDescription('label');
			if (description === undefined || description.longForm === undefined) {
				throw new Error('No edge label property description found.');
			}
			for (let item of description.longForm) {
				this.edgeLabels.set(item[1], item[0] as number);
			}
			// Item edge Compressor
			decompressor = Decompressor.get(metaData.compressors.itemEdgeCompressor);
			if (decompressor === undefined) {
				throw new Error('No item edge decompressor found.');
			}
			description = decompressor.getPropertyDescription('property');
			if (description === undefined || description.longForm === undefined) {
				throw new Error('No item property description found.');
			}
			for (let item of description.longForm) {
				this.itemEdgeProperties.set(item[1], item[0] as number);
			}
		}
	}

	public getProjectRoot(): URI {
		return this.projectRoot;
	}

	public close(): void {
		this.db.close();
	}

	protected getDocumentInfos(): DocumentInfo[] {
		let result = this.allDocumentsStmt.all();
		if (result === undefined) {
			return [];
		}
		return result;
	}

	protected findFile(uri: string): Id | undefined {
		let result = this.findDocumentStmt.get(uri);
		return result;
	}

	protected fileContent(id: Id): string {
		let result: ContentResult = this.getDocumentContentStmt.get(id);
		if (!result || !result.content) {
			return '';
		}
		return Buffer.from(result.content).toString('base64');
	}

	public foldingRanges(uri: string): lsp.FoldingRange[] | undefined {
		let foldingResult = this.getResultForDocument(this.toDatabase(uri), EdgeLabels.textDocument_foldingRange);
		if (foldingResult === undefined) {
			return undefined;
		}
		return foldingResult.result;
	}

	public documentSymbols(uri: string): lsp.DocumentSymbol[] | undefined {
		let symbolResult = this.getResultForDocument(this.toDatabase(uri), EdgeLabels.textDocument_documentSymbol);
		if (symbolResult === undefined) {
			return undefined;
		}
		if (symbolResult.result.length === 0) {
			return [];
		}
		if (lsp.DocumentSymbol.is(symbolResult.result[0])) {
			return symbolResult.result as lsp.DocumentSymbol[];
		} else {
			const vertexRetriever = new VertexRetriever(this.db, 16);
			let collectRanges = (element: RangeBasedDocumentSymbol) => {
				vertexRetriever.add(element.id);
				if (element.children) {
					element.children.forEach(collectRanges);
				}
			}
			let convert = (result: lsp.DocumentSymbol[], elements: RangeBasedDocumentSymbol[], ranges: Map<Id, Range>) => {
				for (let element of elements) {
					let range = ranges.get(element.id);
					if (range !== undefined) {
						let symbol: lsp.DocumentSymbol | undefined = this.asDocumentSymbol(range);
						if (symbol) {
							result.push(symbol);
							if (element.children !== undefined && element.children.length > 0) {
								symbol.children = [];
								convert(symbol.children, element.children, ranges);
							}
						}
					}
				}
			}
			(symbolResult.result as RangeBasedDocumentSymbol[]).forEach(collectRanges);
			let data = vertexRetriever.run();
			let ranges: Map<Id, Range> = new Map();
			for (let element of data) {
				let range: Range = this.decompress(JSON.parse(element.value));
				if (range) {
					ranges.set(range.id, range);
				}
			}
			let result: lsp.DocumentSymbol[] = [];
			convert(result, symbolResult.result as RangeBasedDocumentSymbol[], ranges);
			return result;
		}
	}

	public hover(uri: string, position: lsp.Position): lsp.Hover | undefined {
		let range = this.findRange(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}

		let hoverResult = this.getResultForRange(range.id, EdgeLabels.textDocument_hover);
		if (hoverResult === undefined || hoverResult.result === undefined) {
			return undefined;
		}
		let result: lsp.Hover = Object.assign(Object.create(null), hoverResult.result);
		if (result.range === undefined) {
			result.range = {
				start: {
					line: range.startLine,
					character: range.startCharacter
				},
				end: {
					line: range.endLine,
					character: range.endCharacter
				}
			};
		}
		// Workaround to remove empty object. Need to find out why they are in the dump
		// in the first place.
		if (Array.isArray(result.contents)) {
			for (let i = 0; i < result.contents.length;) {
				let elem = result.contents[i];
				if (typeof elem !== 'string' && elem.language === undefined && elem.value === undefined) {
					result.contents.splice(i, 1);
				} else {
					i++;
				}
			}
		}
		return result;
	}

	public declarations(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
		let range = this.findRange(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}
		let declarationResult = this.getResultForRange(range.id, EdgeLabels.textDocument_declaration);
		if (declarationResult === undefined) {
			return undefined;
		}

		let result: lsp.Location[] = [];
		let queryResult: LocationResult[] = this.findRangeFromResult.all({ id: declarationResult.id });
		if (queryResult && queryResult.length > 0) {
			for(let item of queryResult) {
				result.push(this.createLocation(item));
			}
		}
		return result;
	}

	public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
		let range = this.findRange(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}
		let definitionResult = this.getResultForRange(range.id, EdgeLabels.textDocument_definition);
		if (definitionResult === undefined) {
			return undefined;
		}

		let result: lsp.Location[] = [];
		let queryResult: LocationResult[] = this.findRangeFromResult.all({ id: definitionResult.id });
		if (queryResult && queryResult.length > 0) {
			for(let item of queryResult) {
				result.push(this.createLocation(item));
			}
		}
		return result;
	}

	public references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined {
		let range = this.findRange(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}
		let referenceResult = this.getResultForRange(range.id, EdgeLabels.textDocument_references);
		if (referenceResult === undefined) {
			return undefined;
		}

		let result: lsp.Location[] = [];
		this.resolveReferenceResult(result, referenceResult, context, new Set());
		return result;
	}

	private resolveReferenceResult(locations: lsp.Location[], referenceResult: ReferenceResult, context: lsp.ReferenceContext, dedup: Set<Id>): void {
		let qr: LocationResultWithProperty[] = this.findRangeFromReferenceResult.all({ id: referenceResult.id });
		if (qr && qr.length > 0) {
			let refLabel = this.getItemEdgeProperty(ItemEdgeProperties.references);
			for (let item of qr) {
				if (item.property === refLabel || context.includeDeclaration && !dedup.has(item.id)) {
					dedup.add(item.id);
					locations.push(this.createLocation(item));
				}
			}
		}
		let rqr: VertexResult[] = this.findResultFromReferenceResult.all({ id: referenceResult.id });
		if (rqr && rqr.length > 0) {
			for (let item of rqr) {
				this.resolveReferenceResult(locations, this.decompress(JSON.parse(item.value)), context, dedup);
			}
		}

	}

	private findRange(uri: string, position: lsp.Position): RangeResult | undefined {
		let result: RangeResult[] = this.findRangeStmt.all({ uri: uri, line: position.line, character: position.character});
		if (result === undefined || result.length === 0) {
			return undefined;
		}
		// ToDo we need to sort the result and take the shortest. Since we have a index
		// on the table the shortest one should come last.
		return result[result.length - 1];
	}

	private getResultForRange(rangeId: Id, label: EdgeLabels.textDocument_hover): HoverResult | undefined;
	private getResultForRange(rangeId: Id, label: EdgeLabels.textDocument_declaration): DeclarationResult | undefined;
	private getResultForRange(rangeId: Id, label: EdgeLabels.textDocument_definition): DefinitionResult | undefined;
	private getResultForRange(rangeId: Id, label: EdgeLabels.textDocument_references): ReferenceResult | undefined;
	private getResultForRange(rangeId: Id, label: EdgeLabels): any | undefined {
		let rows = this.findResultStmt.all({ source: rangeId, label: this.getEdgeLabel(label)});
		let result: any | undefined;
		if (rows.length > 0) {
			result = rows[0];
		}
		if (result === undefined) {
			rows = this.findResultViaSetStmt.all({ source: rangeId, label: this.getEdgeLabel(label)});
			if (rows === undefined || rows.length !== 1) {
				return undefined;
			}
			result = rows[0];
		}
		return this.decompress(JSON.parse(result.value));
	}

	private getEdgeLabel(label: EdgeLabels): EdgeLabels | number {
		if (this.edgeLabels === undefined) {
			return label;
		}
		let result = this.edgeLabels.get(label);
		return result !== undefined ? result : label;
	}

	private getItemEdgeProperty(prop: ItemEdgeProperties): ItemEdgeProperties | number {
		if (this.itemEdgeProperties === undefined) {
			return prop;
		}
		let result = this.itemEdgeProperties.get(prop);
		return result !== undefined ? result : prop;
	}

	private createLocation(data: LocationResult): lsp.Location {
		return lsp.Location.create(this.fromDatabase(data.uri), lsp.Range.create(data.startLine, data.startCharacter, data.endLine, data.endCharacter));
	}

	private getResultForDocument(uri: string, label: EdgeLabels.textDocument_documentSymbol): DocumentSymbolResult | undefined;
	private getResultForDocument(uri: string, label: EdgeLabels.textDocument_foldingRange): FoldingRangeResult | undefined;
	private getResultForDocument(uri: string, label: EdgeLabels): any | undefined {
		let data: DocumentResult = this.findResultForDocumentStmt.get({ uri, label: this.getEdgeLabel(label) });
		if (data === undefined) {
			return undefined;
		}
		return this.decompress(JSON.parse(data.value));
	}

	private decompress(value: any): any {
		if (Array.isArray(value)) {
			let decompressor = Decompressor.get(value[0]);
			if (decompressor) {
				return decompressor.decompress(value);
			}
		}
		return value;
	}
}