import { createGunzip } from 'zlib'
import { Readable } from 'stream'

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
