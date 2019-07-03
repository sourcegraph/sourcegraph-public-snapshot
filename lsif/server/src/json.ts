/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
import * as fs from 'fs';
import * as readline from 'readline';

import { URI } from 'vscode-uri';
import * as SemVer from 'semver';

import * as lsp from 'vscode-languageserver';
import {
	Id, Vertex, Project, Document, Range, DiagnosticResult, DocumentSymbolResult, FoldingRangeResult, DocumentLinkResult, DefinitionResult,
	TypeDefinitionResult, HoverResult, ReferenceResult, ImplementationResult, Edge, RangeBasedDocumentSymbol, DeclarationResult, ResultSet,
	ElementTypes, VertexLabels, EdgeLabels, ItemEdgeProperties
} from 'lsif-protocol';

import { DocumentInfo } from './files';
import { Database, UriTransformer } from './database';

interface Vertices {
	all: Map<Id, Vertex>;
	projects: Map<Id, Project>;
	documents: Map<Id, Document>;
	ranges: Map<Id, Range>;
}

type ItemTarget =
	Range |
	{ type: ItemEdgeProperties.declarations; range: Range; } |
	{ type: ItemEdgeProperties.definitions; range: Range; } |
	{ type: ItemEdgeProperties.references; range: Range; } |
	{ type: ItemEdgeProperties.referenceResults; result: ReferenceResult; };

interface Out {
	contains: Map<Id, Document[] | Range[]>;
	item: Map<Id, ItemTarget[]>;
	next: Map<Id, ResultSet>;
	documentSymbol: Map<Id, DocumentSymbolResult>;
	foldingRange: Map<Id, FoldingRangeResult>;
	documentLink: Map<Id, DocumentLinkResult>;
	diagnostic: Map<Id, DiagnosticResult>;
	declaration: Map<Id, DeclarationResult>;
	definition: Map<Id, DefinitionResult>;
	typeDefinition: Map<Id, TypeDefinitionResult>;
	hover: Map<Id, HoverResult>;
	references: Map<Id, ReferenceResult>;
	implementation: Map<Id, ImplementationResult>;
}

interface In {
	contains: Map<Id, Project | Document>;
}

interface Indices {
	documents: Map<string, Document>;
}


export class JsonDatabase extends Database {

	private version: string | undefined;
	private projectRoot!: URI;

	private vertices: Vertices;
	private indices: Indices;
	private out: Out;
	private in: In;

	constructor() {
		super();
		this.vertices = {
			all: new Map(),
			projects: new Map(),
			documents: new Map(),
			ranges: new Map()
		};

		this.indices = {
			documents: new Map()
		};

		this.out = {
			contains: new Map(),
			item: new Map(),
			next: new Map(),
			documentSymbol: new Map(),
			foldingRange: new Map(),
			documentLink: new Map(),
			diagnostic: new Map(),
			declaration: new Map(),
			definition: new Map(),
			typeDefinition: new Map(),
			hover: new Map(),
			references: new Map(),
			implementation: new Map()
		};

		this.in = {
			contains: new Map()
		}
	}

	public load(file: string, transformerFactory: (projectRoot: string) => UriTransformer): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			let input: fs.ReadStream = fs.createReadStream(file, { encoding: 'utf8'});
			const rd = readline.createInterface(input);
			rd.on('line', (line: string) => {
				if (!line || line.length === 0) {
					return;
				}
				let element: Edge | Vertex = JSON.parse(line);
				switch (element.type) {
					case ElementTypes.vertex:
						this.processVertex(element);
						break;
					case ElementTypes.edge:
						this.processEdge(element);
						break;
				}
			});
			rd.on('close', () => {
				if (this.projectRoot === undefined) {
					reject(new Error('No project root provided.'));
					return;
				}
				if (this.version === undefined) {
					reject(new Error('No version found.'));
					return;
				} else {
					let semVer = SemVer.parse(this.version);
					if (!semVer) {
						reject(new Error(`No valid semantic version string. The version is: ${this.version}`));
						return;
					}
					let range: SemVer.Range = new SemVer.Range('>=0.4.0 <0.5.0');
					range.includePrerelease = true;
					if (!SemVer.satisfies(semVer, range)) {
						reject(new Error(`Requires version 0.4.1 but received: ${this.version}`));
						return;
					}
				}
				resolve();
			});

		}).then(() => {
			this.initialize(transformerFactory);
		});
	}

	public getProjectRoot(): URI {
		return this.projectRoot;
	}

	public close(): void {
	}

	private processVertex(vertex: Vertex): void {
		this.vertices.all.set(vertex.id, vertex);
		switch(vertex.label) {
			case VertexLabels.metaData:
				this.version = vertex.version;
				if (vertex.projectRoot !== undefined) {
					this.projectRoot = URI.parse(vertex.projectRoot);
				}
				break;
			case VertexLabels.project:
				this.vertices.projects.set(vertex.id, vertex);
				break;
			case VertexLabels.document:
				this.vertices.documents.set(vertex.id, vertex);
				this.indices.documents.set(vertex.uri, vertex);
				break;
			case VertexLabels.range:
				this.vertices.ranges.set(vertex.id, vertex);
				break;
		}
	}

	private processEdge(edge: Edge): void {
		let property: ItemEdgeProperties | undefined;
		if (edge.label === 'item') {
			property = edge.property;
		}
		if (Edge.is11(edge)) {
			this.doProcessEdge(edge.label, edge.outV, edge.inV, property);
		} else if (Edge.is1N(edge)) {
			for (let inV of edge.inVs) {
				this.doProcessEdge(edge.label, edge.outV, inV, property);
			}
		}
	}

	private doProcessEdge(label: EdgeLabels, outV: Id, inV: Id, property?: ItemEdgeProperties): void {
		let from: Vertex | undefined = this.vertices.all.get(outV);
		let to: Vertex | undefined = this.vertices.all.get(inV);
		if (from === undefined) {
			throw new Error(`No vertex found for Id ${outV}`);
		}
		if (to === undefined) {
			throw new Error(`No vertex found for Id ${inV}`);
		}
		let values: any[] | undefined;
		switch (label) {
			case EdgeLabels.contains:
				values = this.out.contains.get(from.id);
				if (values === undefined) {
					values = [ to as any ];
					this.out.contains.set(from.id, values);
				} else {
					values.push(to);
				}
				this.in.contains.set(to.id, from as any);
				break;
			case EdgeLabels.item:
				values = this.out.item.get(from.id);
				let itemTarget: ItemTarget | undefined;
				if (property !== undefined) {
					switch (property) {
						case ItemEdgeProperties.references:
							itemTarget = { type: property, range: to as Range };
							break;
						case ItemEdgeProperties.declarations:
							itemTarget = { type: property, range: to as Range };
							break;
						case ItemEdgeProperties.definitions:
							itemTarget = { type: property, range: to as Range };
							break;
						case ItemEdgeProperties.referenceResults:
							itemTarget = { type: property, result: to as ReferenceResult };
							break;
					}
				} else {
					itemTarget = to as Range;
				}
				if (itemTarget !== undefined) {
					if (values === undefined) {
						values = [ itemTarget ];
						this.out.item.set(from.id, values);
					} else {
						values.push(itemTarget);
					}
				}
				break;
			case EdgeLabels.next:
				this.out.next.set(from.id, to as ResultSet);
				break;
			case EdgeLabels.textDocument_documentSymbol:
				this.out.documentSymbol.set(from.id, to as DocumentSymbolResult);
				break;
			case EdgeLabels.textDocument_foldingRange:
				this.out.foldingRange.set(from.id, to as FoldingRangeResult);
				break;
			case EdgeLabels.textDocument_documentLink:
				this.out.documentLink.set(from.id, to as DocumentLinkResult);
				break;
			case EdgeLabels.textDocument_diagnostic:
				this.out.diagnostic.set(from.id, to as DiagnosticResult);
				break;
			case EdgeLabels.textDocument_definition:
				this.out.definition.set(from.id, to as DefinitionResult);
				break;
			case EdgeLabels.textDocument_typeDefinition:
				this.out.typeDefinition.set(from.id, to as TypeDefinitionResult);
				break;
			case EdgeLabels.textDocument_hover:
				this.out.hover.set(from.id, to as HoverResult);
				break;
			case EdgeLabels.textDocument_references:
				this.out.references.set(from.id, to as ReferenceResult);
				break;
		}
	}

	public getDocumentInfos(): DocumentInfo[] {
		let result: DocumentInfo[] = [];
		this.vertices.documents.forEach((document, key) => {
			result.push({ uri: document.uri, id: key });
		});
		return result;
	}

	protected findFile(uri: string): Id | undefined {
		let result = this.indices.documents.get(uri);
		if (result == undefined) {
			return undefined;
		}
		return result.id;
	}

	protected fileContent(id: Id): string | undefined {
		let document = this.vertices.documents.get(id);
		if (document === undefined) {
			return undefined;
		}
		return document.contents;
	}

	public foldingRanges(uri: string): lsp.FoldingRange[] | undefined {
		let document = this.indices.documents.get(this.toDatabase(uri));
		if (document === undefined) {
			return undefined;
		}
		let foldingRangeResult = this.out.foldingRange.get(document.id);
		if (foldingRangeResult === undefined) {
			return undefined;
		}
		let result: lsp.FoldingRange[] = [];
		for (let item of foldingRangeResult.result) {
			result.push(Object.assign(Object.create(null), item));
		}
		return result;
	}

	public documentSymbols(uri: string): lsp.DocumentSymbol[] | undefined {
		let document = this.indices.documents.get(this.toDatabase(uri));
		if (document === undefined) {
			return undefined;
		}
		let documentSymbolResult = this.out.documentSymbol.get(document.id);
		if (documentSymbolResult === undefined || documentSymbolResult.result.length === 0) {
			return undefined;
		}
		let first = documentSymbolResult.result[0];
		let result: lsp.DocumentSymbol[] = [];
		if (lsp.DocumentSymbol.is(first)) {
			for (let item of documentSymbolResult.result) {
				result.push(Object.assign(Object.create(null), item));
			}
		} else {
			for (let item of (documentSymbolResult.result as RangeBasedDocumentSymbol[])) {
				let converted = this.toDocumentSymbol(item);
				if (converted !== undefined) {
					result.push(converted);
				}
			}
		}
		return result;
	}

	private toDocumentSymbol(value: RangeBasedDocumentSymbol): lsp.DocumentSymbol | undefined {
		let range = this.vertices.ranges.get(value.id)!;
		let tag = range.tag;
		if (tag === undefined || !(tag.type === 'declaration' || tag.type === 'definition')) {
			return undefined;
		}
		let result: lsp.DocumentSymbol = lsp.DocumentSymbol.create(
			tag.text, tag.detail || '', tag.kind,
			tag.fullRange, this.asRange(range)
		)
		if (value.children && value.children.length > 0) {
			result.children = [];
			for (let child of value.children) {
				let converted = this.toDocumentSymbol(child);
				if (converted !== undefined) {
					result.children.push(converted);
				}
			}
		}
		return result;
	}

	public hover(uri: string, position: lsp.Position): lsp.Hover | undefined {
		let range = this.findRangeFromPosition(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}

		let hoverResult: HoverResult | undefined = this.getResult(range, this.out.hover);
		if (hoverResult === undefined) {
			return undefined;
		}

		let hoverRange = hoverResult.result.range !== undefined ? hoverResult.result.range : range;
		return {
			contents: hoverResult.result.contents,
			range: hoverRange
		};
	}

	public declarations(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
		let range = this.findRangeFromPosition(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}
		let declarationResult: DeclarationResult | undefined = this.getResult(range, this.out.declaration);
		if (declarationResult === undefined) {
			return undefined;
		}
		let ranges = this.item(declarationResult);
		if (ranges === undefined) {
			return undefined;
		}
		let result: lsp.Location[] = [];
		for (let element of ranges) {
			result.push(this.asLocation(element));
		}
		return result;
	}

	public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
		let range = this.findRangeFromPosition(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}
		let definitionResult: DefinitionResult | undefined = this.getResult(range, this.out.definition);
		if (definitionResult === undefined) {
			return undefined;
		}
		let ranges = this.item(definitionResult);
		if (ranges === undefined) {
			return undefined;
		}
		let result: lsp.Location[] = [];
		for (let element of ranges) {
			result.push(this.asLocation(element));
		}
		return result;
	}

	public references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined {
		let range = this.findRangeFromPosition(this.toDatabase(uri), position);
		if (range === undefined) {
			return undefined;
		}

		let referenceResult: ReferenceResult | undefined = this.getResult(range, this.out.references);
		if (referenceResult === undefined) {
			return undefined;
		}

		let targets = this.item(referenceResult);
		if (targets === undefined) {
			return undefined;
		}
		return this.asReferenceResult(targets, context, new Set());
	}

	private getResult<T>(range: Range, edges: Map<Id, T>): T | undefined {
		let id: Id | undefined = range.id;
		do {
			let result: T | undefined = edges.get(id);
			if (result !== undefined) {
				return result;
			}
			let next = this.out.next.get(id);
			id = next !== undefined ? next.id : undefined;
		} while (id !== undefined);
		return undefined;
	}

	private item(value: DeclarationResult): Range[];
	private item(value: DefinitionResult): Range[];
	private item(value: ReferenceResult): ItemTarget[];
	private item(value: DeclarationResult | DefinitionResult | ReferenceResult): Range[] | ItemTarget[] | undefined {
		if (value.label === 'declarationResult') {
			return this.out.item.get(value.id) as Range[];
		} else if (value.label === 'definitionResult') {
			return this.out.item.get(value.id) as Range[];
		} else if (value.label === 'referenceResult') {
			return this.out.item.get(value.id) as ItemTarget[];
		} else {
			return undefined;
		}
	}

	private asReferenceResult(targets: ItemTarget[], context: lsp.ReferenceContext, dedup: Set<Id>): lsp.Location[] {
		let result: lsp.Location[] = [];
		for (let target of targets) {
			if (target.type === ItemEdgeProperties.declarations && context.includeDeclaration) {
				this.addLocation(result, target.range, dedup);
			} else if (target.type === ItemEdgeProperties.definitions && context.includeDeclaration) {
				this.addLocation(result, target.range, dedup);
			} else if (target.type === ItemEdgeProperties.references) {
				this.addLocation(result, target.range, dedup);
			} else if (target.type === ItemEdgeProperties.referenceResults) {
				result.push(...this.asReferenceResult(this.item(target.result), context, dedup));
			}
		}
		return result;
	}

	private addLocation(result: lsp.Location[], value: Range | lsp.Location, dedup: Set<Id>): void {
		if (lsp.Location.is(value)) {
			result.push(value);
		} else {
			if (dedup.has(value.id)) {
				return;
			}
			let document = this.in.contains.get(value.id)!;
			result.push(lsp.Location.create(this.fromDatabase((document as Document).uri), this.asRange(value)));
			dedup.add(value.id);
		}
	}

	private findRangeFromPosition(file: string, position: lsp.Position): Range | undefined {
		let document = this.indices.documents.get(file);
		if (document === undefined) {
			return undefined;
		}
		let contains = this.out.contains.get(document.id);
		if (contains === undefined || contains.length === 0) {
			return undefined;
		}

		let candidate: Range | undefined;
		for (let item of contains) {
			if (item.label !== VertexLabels.range) {
				continue;
			}
			if (JsonDatabase.containsPosition(item, position)) {
				if (!candidate) {
					candidate = item;
				} else {
					if (JsonDatabase.containsRange(candidate, item)) {
						candidate = item;
					}
				}
			}
		}
		return candidate;
	}

	private asLocation(value: Range | lsp.Location): lsp.Location {
		if (lsp.Location.is(value)) {
			return value;
		} else {
			let document = this.in.contains.get(value.id)!;
			return lsp.Location.create(this.fromDatabase((document as Document).uri), this.asRange(value));
		}
	}

	private static containsPosition(range: lsp.Range, position: lsp.Position): boolean {
		if (position.line < range.start.line || position.line > range.end.line) {
			return false;
		}
		if (position.line === range.start.line && position.character < range.start.character) {
			return false;
		}
		if (position.line === range.end.line && position.character > range.end.character) {
			return false;
		}
		return true;
	}

	/**
	 * Test if `otherRange` is in `range`. If the ranges are equal, will return true.
	 */
	public static containsRange(range: lsp.Range, otherRange: lsp.Range): boolean {
		if (otherRange.start.line < range.start.line || otherRange.end.line < range.start.line) {
			return false;
		}
		if (otherRange.start.line > range.end.line || otherRange.end.line > range.end.line) {
			return false;
		}
		if (otherRange.start.line === range.start.line && otherRange.start.character < range.start.character) {
			return false;
		}
		if (otherRange.end.line === range.end.line && otherRange.end.character > range.end.character) {
			return false;
		}
		return true;
	}
}
