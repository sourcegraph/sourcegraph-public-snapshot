/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as fs from 'fs'
import * as readline from 'readline'
import { Edge, Vertex, ElementTypes, VertexLabels } from 'lsif-protocol'
import { CompressorPropertyDescription, MetaData } from './protocol.compress'
import {
    Compressor,
    CompressorProperty,
    vertexShortForms,
    edgeShortForms,
    vertexCompressor,
    edge11Compressor,
    itemEdgeCompressor,
} from './compress'
import * as graph from './graphStore'
import * as blob from './blobStore'

export async function convertToGraph(inFile: string, outFile: string): Promise<void> {
    return await convertToStore(inFile, new graph.GraphStore(outFile, stringify, shortForm))
}

export async function convertToBlob(inFile: string, outFile: string): Promise<void> {
    return await convertToStore(inFile, new blob.BlobStore(outFile, '0', true))
}

function stringify(element: Vertex | Edge): string {
    if (element.type === ElementTypes.vertex && element.label === VertexLabels.metaData) {
        return JSON.stringify(element, undefined, 0)
    }

    const compressor = Compressor.getCompressor(element)
    if (compressor === undefined) {
        throw new Error(`No compressor found for ${element.label}`)
    }

    return JSON.stringify(compressor.compress(element, { mode: 'store' }))
}

function shortForm(element: Vertex | Edge): number {
    const result =
        element.type === ElementTypes.vertex ? vertexShortForms.get(element.label) : edgeShortForms.get(element.label)

    if (result === undefined) {
        throw new Error(`Can't compute short form for ${element.label}`)
    }

    return result
}

function convertToStore(inFile: string, db: graph.GraphStore | blob.BlobStore): Promise<void> {
    const input = fs.createReadStream(inFile, { encoding: 'utf8' })

    const insertElement = (element: Vertex | Edge): void => {
        if (element.type === ElementTypes.vertex && element.label === VertexLabels.metaData) {
            insertMetadata(element)
        }

        db.insert(element)
    }

    const insertMetadata = (element: MetaData): void => {
        const convertMetaData = (data: CompressorProperty): CompressorPropertyDescription => {
            const result: CompressorPropertyDescription = {
                name: data.name as string,
                index: data.index,
                compressionKind: data.compressionKind,
            }

            if (data.shortForm === undefined) {
                return result
            }

            const long: Set<string> = new Set()
            const short: Set<string | number> = new Set()
            result.shortForm = []

            for (const [key, value] of data.shortForm) {
                if (long.has(key)) {
                    throw new Error(`Duplicate key ${key} in short form.`)
                }

                if (short.has(value)) {
                    throw new Error(`Duplicate value ${value} in short form.`)
                }

                long.add(key)
                short.add(value)
                result.shortForm.push([key, value])
            }

            return result
        }

        const compressors = Compressor.allCompressors()
        if (compressors.length > 0) {
            const compressMetaData: MetaData = element as MetaData
            compressMetaData.compressors = {
                vertexCompressor: vertexCompressor.id,
                edgeCompressor: edge11Compressor.id,
                itemEdgeCompressor: itemEdgeCompressor.id,
                all: [],
            }

            for (const compressor of compressors) {
                compressMetaData.compressors.all.push({
                    id: compressor.id,
                    parent: compressor.parent !== undefined ? compressor.parent.id : undefined,
                    properties: compressor.metaData().map(convertMetaData),
                })
            }
        }
    }

    return new Promise((resolve, reject) => {
        db.runInsertTransaction(() => {
            const rd = readline.createInterface(input)
            rd.on('line', line => {
                if (!line) {
                    return
                }

                let element: Edge | Vertex
                try {
                    element = JSON.parse(line)
                } catch (err) {
                    throw new Error(`Parsing failed for line:\n${line}`)
                }

                try {
                    insertElement(element)
                } catch (e) {
                    throw e
                }
            })

            rd.on('close', () => {
                db.close()
                resolve()
            })
        })
    })
}
