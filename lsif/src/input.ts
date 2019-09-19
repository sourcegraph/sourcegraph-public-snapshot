import * as zlib from 'mz/zlib'
import { Edge, Vertex } from 'lsif-protocol'
import { Readable, Writable } from 'stream'
import { ValidateFunction } from 'ajv'
import { readline } from 'mz'

/**
 * Reads the input and writes it to the output stream unchanged. The input
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

    return new Promise((resolve, reject) => {
        const gunzippedReader = input.pipe(zlib.createGunzip()).on('error', reject)

        const gzipWriter = zlib
            .createGzip()
            .pipe(output)
            .on('error', reject)
            .on('finish', resolve)

        processElements(readline.createInterface({ input: gunzippedReader }), (element, data) => {
            if (!validator(element) && validator.errors) {
                throw new Error(validator.errors.map(e => e.message).join(', '))
            }

            gzipWriter.write(data)
            gzipWriter.write('\n')
        }).then(() => gzipWriter.end(), reject)
    })
}

/**
 * Read the input and call a function on the parsed JSON element on each line.
 * The input is expected to be a gzipped sequence of JSON lines.
 *
 * @param input The gzipped input stream.
 * @param process The function to apply to element of the input stream.
 */
export function processLsifInput(input: Readable, process: (element: Vertex | Edge) => void): Promise<void> {
    return new Promise((resolve, reject) => {
        const gunzippedReader = input.pipe(zlib.createGunzip()).on('error', reject)
        processElements(readline.createInterface({ input: gunzippedReader }), process).then(resolve, reject)
    })
}

/**
 * Invoke a callback with the parsed JSON element from each line of the input.
 *
 * @param lines An async iterator of JSOn lines.
 * @param process The function to invoke on each parsed JSON element.
 */
async function processElements(
    lines: AsyncIterable<string>,
    process: (element: Vertex | Edge, data: string) => void
): Promise<void> {
    let line = 0
    for await (const data of lines) {
        line++

        try {
            process(JSON.parse(data), data)
        } catch (e) {
            throw new Error(`Failed to validate line #${line} (${data}): ${e && e.message}`)
        }
    }
}
