/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
import { URI } from 'vscode-uri';
import * as lsp from 'vscode-languageserver';
import { Range, Id } from 'lsif-protocol';

import { FileType, FileSystem, DocumentInfo, FileStat } from './files';

export interface UriTransformer {
	toDatabase(uri: string): string;
	fromDatabase(uri: string): string;
}

export const noopTransformer: UriTransformer = {
	toDatabase: uri => uri,
	fromDatabase: uri => uri
}

export abstract class Database {

	private fileSystem!: FileSystem;
	private uriTransformer!: UriTransformer;

	protected constructor() {
	}

	protected initialize(transformerFactory: (projectRoot: string) => UriTransformer): void {
		const projectRoot = this.getProjectRoot().toString(true);
		this.uriTransformer = transformerFactory ? transformerFactory(projectRoot) : noopTransformer;
		this.fileSystem = new FileSystem(projectRoot, this.getDocumentInfos());
	}

	public abstract load(file: string, transformerFactory: (projectRoot: string) => UriTransformer): Promise<void>;

	public abstract close(): void;

	public abstract getProjectRoot(): URI;

	protected abstract getDocumentInfos(): DocumentInfo[];

	public stat(uri: string): FileStat | null {
		let transformed = this.uriTransformer.toDatabase(uri);
		let result = this.fileSystem.stat(transformed);
		if (result !== null) {
			return result;
		}
		let id = this.findFile(transformed);
		if (id === undefined) {
			return null;
		}
		return FileStat.createFile();
	}

	public readDirectory(uri: string): [string, FileType][] {
		return this.fileSystem.readDirectory(this.uriTransformer.toDatabase(uri));
	}

	public readFileContent(uri: string): string | null {
		let transformed = this.uriTransformer.toDatabase(uri);
		let id = this.fileSystem.getFileId(transformed);
		if (id === undefined) {
			id = this.findFile(transformed);
		}
		if (id === undefined) {
			return null;
		}
		let result = this.fileContent(id);
		if (result === undefined) {
			return null;
		}
		return result;
	}

	protected abstract findFile(uri: string): Id | undefined;

	protected abstract fileContent(id: Id): string | undefined;

	public abstract foldingRanges(uri: string): lsp.FoldingRange[] | undefined;

	public abstract documentSymbols(uri: string): lsp.DocumentSymbol[] | undefined;

	public abstract hover(uri: string, position: lsp.Position): lsp.Hover | undefined;

	public abstract declarations(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined;

	public abstract definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined;

	public abstract references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined;

	protected asDocumentSymbol(range: Range): lsp.DocumentSymbol | undefined {
		let tag = range.tag;
		if (tag === undefined || !(tag.type === 'declaration' || tag.type === 'definition')) {
			return undefined;
		}
		return lsp.DocumentSymbol.create(
			tag.text, tag.detail || '', tag.kind,
			tag.fullRange, this.asRange(range)
		)
	}

	protected asRange(value: Range): lsp.Range {
		return {
			start: {
				line: value.start.line,
				character: value.start.character
			},
			end: {
				line: value.end.line,
				character: value.end.character
			}
		};
	}

	protected toDatabase(uri: string): string {
		return this.uriTransformer.toDatabase(uri);
	}

	protected fromDatabase(uri: string): string {
		return this.uriTransformer.fromDatabase(uri);
	}
}
