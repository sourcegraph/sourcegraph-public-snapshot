import * as zlib from 'mz/zlib'
import { Edge, Vertex } from 'lsif-protocol'
import { Readable, Writable } from 'stream'
import { ValidateFunction } from 'ajv'
import { readline } from 'mz'

/**
 * Pipes the input data into the output stream. All content of the input
 * stream will be written to the output stream if there are no validation
 * errors. The input is assumed to be gzipped. If the validator is
 * undefined, no validation will occur.
 *
 * @param input The input stream.
 * @param output The output stream.
 * @param validator The JSON schema validation function to apply.
 */
export async function validateLsifInput(
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

    const gzipWriter = zlib.createGzip()
    const promise = new Promise((resolve, reject) => {
        gzipWriter
            .pipe(output)
            .on('error', reject)
            .on('finish', resolve)
    })

    let line = 0
    for await (const data of readline.createInterface({ input: input.pipe(zlib.createGunzip()) })) {
        line++

        try {
            if (!validator(JSON.parse(data)) && validator.errors) {
                throw new Error(validator.errors.map(e => e.message).join(', '))
            }

            gzipWriter.write(data)
        } catch (e) {
            throw new Error(`Failed to validate line #${line} (${data}): ${e && e.message}`)
        }
    }

    await promise
}

/**
 * Apply a function to each element of the input stream.
 *
 * @param input The gzipped input stream.
 * @param process The function to apply to element of the input stream.
 */
export async function processLsifInput(input: Readable, process: (element: Vertex | Edge) => void): Promise<void> {
    let line = 0
    for await (const data of readline.createInterface({ input: input.pipe(zlib.createGunzip()) })) {
        line++

        try {
            process(JSON.parse(data))
        } catch (e) {
            throw new Error(`Failed to process line #${line} (${data}): ${e && e.message}`)
        }
    }
}
