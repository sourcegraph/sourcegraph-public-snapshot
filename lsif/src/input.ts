import * as definitionsSchema from './lsif.schema.json'
import Ajv from 'ajv'
import { Edge, Vertex } from 'lsif-protocol'
import { Readable } from 'stream'
import { createGunzip } from 'zlib'

/**
 * A JSON schema validation function that accepts an LSIF vertex or edge.
 */
const lsifElementValidator = new Ajv().addSchema({ $id: 'defs.json', ...definitionsSchema }).compile({
    oneOf: [{ $ref: 'defs.json#/definitions/Vertex' }, { $ref: 'defs.json#/definitions/Edge' }],
})

/**
 * Yield parsed JSON elements from a stream containing the gzipped JSON lines.
 *
 * @param input A stream of gzipped JSON lines.
 */
export async function* readGzippedJsonElements(input: Readable): AsyncIterable<unknown> {
    for await (const element of parseJsonLines(splitLines(input.pipe(createGunzip())))) {
        yield element
    }
}

/**
 * Reads the input stream of parsed LSIF lines and validates it using JSON schema.
 * If it is not a valid vertex or edge, an error is thrown with the line index and
 * content as context. Yields the validated items.
 *
 * @param parsedLines The parsed JSON lines.
 */
export async function* validateLsifElements(parsedLines: AsyncIterable<unknown>): AsyncIterable<Edge | Vertex> {
    let index = 0
    for await (const element of parsedLines) {
        index++

        if (!lsifElementValidator(element) && lsifElementValidator.errors) {
            // TODO - schema messages are not good due to oneOf
            // only take the first error for now to give the user
            // something to work with.
            throw Object.assign(
                new Error(
                    `Invalid LSIF element at index #${index} (${JSON.stringify(element)}): ${
                        lsifElementValidator.errors[0].message
                    }`
                ),
                { element, index }
            )
        }

        yield element as Vertex | Edge
    }
}

/**
 * Transform an async iterable into an async iterable of lines. Each value
 * is stripped of its trailing newline. Lines may be empty.
 *
 * @param input The input buffer.
 */
export async function* splitLines(input: AsyncIterable<string | Buffer>): AsyncIterable<string> {
    let buffer = ''
    for await (const data of input) {
        buffer += data.toString()

        do {
            const index = buffer.indexOf('\n')
            if (index < 0) {
                break
            }

            yield buffer.substring(0, index)
            buffer = buffer.substring(index + 1)
        } while (true)
    }

    yield buffer
}

/**
 * JSON stringifies an iterable of objects and yields them with trailing newlines.
 *
 * @param elements The iterable of objects to stringify.
 */
export async function* stringifyJsonLines(elements: AsyncIterable<unknown>): AsyncIterable<string> {
    for await (const element of elements) {
        yield JSON.stringify(element) + '\n'
    }
}

/**
 * Parses a stream of uncompressed JSON strings and yields each parsed line.
 * Ignores empty lines. Throws an exception with line index and content when
 * a non-empty line is not valid JSON.
 *
 * @param lines An iterable of JSON lines.
 */
export async function* parseJsonLines(lines: AsyncIterable<string>): AsyncIterable<any> {
    let index = 0
    for await (const data of lines) {
        index++

        if (!data) {
            continue
        }

        try {
            yield JSON.parse(data)
        } catch (e) {
            throw new Error(`Failed to process line #${index} (${data}): ${e && e.message}`)
        }
    }
}
