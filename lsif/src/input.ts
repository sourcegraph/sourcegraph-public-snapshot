import * as definitionsSchema from './lsif.schema.json'
import * as zlib from 'mz/zlib'
import Ajv, { ValidateFunction } from 'ajv'
import { Edge, Vertex } from 'lsif-protocol'
import { Readable, Writable } from 'stream'

/**
 * A JSON schema validation function that accepts an LSIF vertex or edge.
 */
export const elementValidator = new Ajv().addSchema({ $id: 'defs.json', ...definitionsSchema }).compile({
    oneOf: [{ $ref: 'defs.json#/definitions/Vertex' }, { $ref: 'defs.json#/definitions/Edge' }],
})

/**
 * Reads the input stream and writes it to the output stream unchanged. The input
 * is expected to be a gzipped sequence of JSON lines. If a validation function
 * is supplied, it will be called with the parsed JSON element of each line.
 *
 * @param input The input stream.
 * @param output The output stream.
 * @param validator The JSON schema validation function to apply.
 */
export function validateLsifInput(
    input: Readable,
    output: Writable,
    validator: ValidateFunction | undefined
): Promise<void> {
    if (!validator) {
        // Not validating, pipe input to temp file
        return new Promise((resolve, reject) => {
            input
                .pipe(output)
                .on('error', reject)
                .on('finish', resolve)
        })
    }

    const lines = processElements(splitLines(input.pipe(zlib.createGunzip())), element => {
        if (!validator(element) && validator.errors) {
            // TODO - schema messages are not good due to oneOf
            // only take the first error for now to give the user
            // something to work with.
            throw new Error(validator.errors[0].message)
        }
    })

    return new Promise((resolve, reject) => {
        Readable.from(addNewlines(lines))
            .on('error', reject)
            .pipe(zlib.createGzip())
            .on('error', reject)
            .pipe(output)
            .on('error', reject)
            .on('finish', resolve)
    })
}

/**
 * Read the input stream and call a function on the parsed JSON element on each
 * line. The input is expected to be a gzipped sequence of JSON lines.
 *
 * @param input The gzipped input stream.
 * @param process The function to apply to element of the input stream.
 */
export async function processLsifInput(input: Readable, process: (element: Vertex | Edge) => void): Promise<void> {
    for await (const _ of processElements(splitLines(input.pipe(zlib.createGunzip())), process)) {
        // no-op body, just consusme the iterable
    }
}

/**
 * Transform an async iterable into an async iterable of lines. Each value
 * is stripped of its trailing newline. Lines may be empty.
 *
 * @param input The input buffer.
 */
export async function* splitLines(input: AsyncIterable<string>): AsyncIterable<string> {
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
 * Add newlines back into the output from `splitLines`.
 *
 * @param lines An iterable of lines.
 */
async function* addNewlines(lines: AsyncIterable<string>): AsyncIterable<string> {
    for await (const line of lines) {
        yield line
        yield '\n'
    }
}

/**
 * Invoke a process function for each parsed JSON element in the input iterable.
 * Wraps errors thrown by the process function with the line data and index
 * context. Does not invoke the process function for empty lines. Returns a copy
 * of the input iterable.
 *
 * @param lines An iterable of JSON lines.
 * @param process The function to invoke for each element.
 */
export async function* processElements(
    lines: AsyncIterable<string>,
    process: (element: Vertex | Edge) => void
): AsyncIterable<string> {
    let line = 0
    for await (const data of lines) {
        line++

        if (data === '') {
            continue
        }

        try {
            process(JSON.parse(data))
            yield data
        } catch (e) {
            throw new Error(`Failed to process line #${line} (${data}): ${e && e.message}`)
        }
    }
}
