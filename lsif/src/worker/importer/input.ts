import * as fs from 'mz/fs'
import { createGunzip } from 'zlib'

/**
 * Yield parsed JSON elements from a file containing the gzipped JSON lines.
 *
 * @param path The filepath containing a gzipped compressed stream of JSON lines composing the LSIF dump.
 */
export function readGzippedJsonElementsFromFile(path: string): AsyncIterable<unknown> {
    const input = fs.createReadStream(path)
    const piped = input.pipe(createGunzip())

    // If we get an error opening or reading the file, we need to ensure that
    // we forward the error to the readable stream that splitLines is consuming.
    // If we don't register this error handler in the same tick as the call to
    // createReadStream, we may miss the error. This uncaught error causes the
    // process to crash.
    //
    // We should obviously like to catch this error instead and fail the single
    // job for which that file is missing.
    //
    // This is a source of problems for others as well:
    // https://stackoverflow.com/questions/17136536/is-enoent-from-fs-createreadstream-uncatchable

    input.on('error', error => piped.emit('error', error))

    // Create the iterable
    return parseJsonLines(splitLines(piped))
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
 * Parses a stream of uncompressed JSON strings and yields each parsed line.
 * Ignores empty lines. Throws an exception with line index and content when
 * a non-empty line is not valid JSON.
 *
 * @param lines An iterable of JSON lines.
 */
export async function* parseJsonLines(lines: AsyncIterable<string>): AsyncIterable<unknown> {
    let index = 0
    for await (const data of lines) {
        index++

        if (!data) {
            continue
        }

        try {
            yield JSON.parse(data)
        } catch (error) {
            throw new Error(`Failed to process line #${index} (${data}): ${error && error.message}`)
        }
    }
}
