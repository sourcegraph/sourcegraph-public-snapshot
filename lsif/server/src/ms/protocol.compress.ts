/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
'use strict';

import * as protocol from 'lsif-protocol';

export enum CompressionKind {
	scalar = 'scalar',
	literal = 'literal',
	array = 'array',
	any = 'any',
	raw = 'raw'
}

export interface CompressorPropertyDescription {
	/**
	 * The name of the property.
	 */
	name: string;

	/**
	 * It's index in the array.
	 */
	index: number;

	/**
	 * Whether the value is raw in case it was an object literal.
	 */
	compressionKind: CompressionKind;

	/**
	 * Short form if the value is a string.
	 */
	shortForm?: [string, string | number][];

}

export interface CompressorData {
	vertexCompressor: number;
	edgeCompressor: number;
	itemEdgeCompressor: number;
	all: CompressorDescription[];
}

export interface CompressorDescription {
	/**
	 * The compressor id.
	 */
	id: number;

	/**
	 * The parent compressor or undefined.
	 */
	parent: number | undefined;

	/**
	 * The compressed propeties.
	 */
	properties: CompressorPropertyDescription[];
}

/**
 * The meta data vertex.
 */
export interface MetaData extends protocol.MetaData {

	/**
	 * A description of the compressor used.
	 */
	compressors?: CompressorData;
}