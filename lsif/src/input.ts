import * as es from 'event-stream'
import * as zlib from 'mz/zlib'
import { ValidateFunction } from 'ajv'
import { Readable, Writable } from 'stream'
import { Vertex, Edge } from 'lsif-protocol'

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
export function validateLsifInput(
    input: Readable,
    output: Writable,
    validator: ValidateFunction | undefined
): Promise<void> {
    if (!validator) {
        return new Promise((resolve, reject) => {
            // If we're not validating, just pipe input to the temp file
            input
                .pipe(output)
                .on('error', reject)
                .on('finish', resolve)
        })
    }

    let line = 0

    // Must check each line synchronously
    // eslint-disable-next-line no-sync
    const lineMapper = es.mapSync((text: string): string => {
        line++
        if (text === '') {
            return text
        }

        if (!validator(parseLine(text, line))) {
            throw new Error(
                `Failed to validate line #${line} (${JSON.stringify(
                    text
                )}): Does not match a known vertex or edge shape.`
            )
        }

        return text
    })

    return new Promise((resolve, reject) => {
        input
            .pipe(zlib.createGunzip())
            .on('error', reject)
            .pipe(es.split())
            .pipe(lineMapper)
            .on('error', reject)
            .pipe(es.join('\n'))
            .pipe(zlib.createGzip())
            .on('error', reject)
            .pipe(output)
            .on('finish', resolve)
    })
}

/**
 * Apply a function to each element of the input stream. The input is
 * assumed to be gzipped.
 *
 * @param input The input stream.
 * @param process The function to apply to element of the input stream.
 */
export function processLsifInput(input: Readable, process: (element: Vertex | Edge) => void): Promise<void> {
    let line = 0

    // Must check each line synchronously
    // eslint-disable-next-line no-sync
    const lineMapper = es.mapSync((text: string): void => {
        line++
        if (text === '') {
            return
        }

        const data = parseLine(text, line)

        try {
            process(data)
        } catch (e) {
            throw new Error(`Failed to process line #${line} (${JSON.stringify(text)}): ${e && e.message}.`)
        }
    })

    return new Promise((resolve, reject) => {
        input
            .pipe(zlib.createGunzip())
            .on('error', reject)
            .pipe(es.split())
            .pipe(lineMapper)
            .on('error', reject)
            .on('end', resolve)
    })
}

/**
 * Parse a single line of an LSIF dump input.
 *
 * @param text The text to parse.
 * @param line The line index (one-based).
 */
function parseLine(text: string, line: number): Vertex | Edge {
    try {
        return JSON.parse(text)
    } catch (e) {
        throw Object.assign(new Error(`Failed to process line #${line} (${JSON.stringify(text)}): Invalid JSON.`))
    }
}
