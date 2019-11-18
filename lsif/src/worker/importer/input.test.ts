import * as fs from 'mz/fs'
import * as path from 'path'
import * as zlib from 'mz/zlib'
import rmfr from 'rmfr'
import { parseJsonLines, readGzippedJsonElementsFromFile, splitLines } from './input'
import { Readable } from 'stream'

describe('readGzippedJsonElements', () => {
    let tempPath!: string

    beforeAll(async () => {
        tempPath = await fs.mkdtemp('test-', { encoding: 'utf8' })
    })

    afterAll(async () => {
        await rmfr(tempPath)
    })

    it('should decode gzip', async () => {
        const lines = [
            { type: 'vertex', label: 'project' },
            { type: 'vertex', label: 'document' },
            { type: 'edge', label: 'item' },
            { type: 'edge', label: 'moniker' },
        ]

        const filename = path.join(tempPath, 'gzip.txt')

        const chunks = []
        for await (const chunk of Readable.from(lines.map(l => JSON.stringify(l)).join('\n')).pipe(zlib.createGzip())) {
            chunks.push(chunk)
        }

        await fs.writeFile(filename, Buffer.concat(chunks))

        const elements: unknown[] = []
        for await (const element of readGzippedJsonElementsFromFile(filename)) {
            elements.push(element)
        }

        expect(elements).toEqual(lines)
    })

    it('should fail without gzip', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const filename = path.join(tempPath, 'nogzip.txt')
        await fs.writeFile(filename, lines.join('\n'))

        await expect(consume(readGzippedJsonElementsFromFile(filename))).rejects.toThrowError(
            new Error('incorrect header check')
        )
    })

    it('should throw an error on IO error', async () => {
        const filename = path.join(tempPath, 'missing.txt')

        await expect(consume(readGzippedJsonElementsFromFile(filename))).rejects.toThrowError(
            new Error(`ENOENT: no such file or directory, open '${filename}'`)
        )
    })
})

describe('splitLines', () => {
    it('should split input by newline', async () => {
        const chunks = ['foo\n', 'bar', '\nbaz\n\nbonk\nqu', 'ux']

        const lines: string[] = []
        for await (const line of splitLines(generate(chunks))) {
            lines.push(line)
        }

        expect(lines).toEqual(['foo', 'bar', 'baz', '', 'bonk', 'quux'])
    })
})

describe('parseJsonLines', () => {
    it('should parse JSON', async () => {
        const lines = [
            { type: 'vertex', label: 'project' },
            { type: 'vertex', label: 'document' },
            { type: 'edge', label: 'item' },
            { type: 'edge', label: 'moniker' },
        ]

        const elements: unknown[] = []
        for await (const element of parseJsonLines(generate(lines.map(l => JSON.stringify(l))))) {
            elements.push(element)
        }

        expect(elements).toEqual(lines)
    })

    it('should wrap parse errors', async () => {
        const input = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge" "label": "moniker"}', // missing comma
        ]

        await expect(consume(parseJsonLines(generate(input)))).rejects.toThrowError(
            new Error(
                'Failed to process line #4 ({"type": "edge" "label": "moniker"}): Unexpected string in JSON at position 16'
            )
        )
    })
})

//
// Helpers

async function* generate<T>(values: T[]): AsyncIterable<T> {
    // Make it actually async
    await Promise.resolve()

    for (const value of values) {
        yield value
    }
}

async function consume(iterable: AsyncIterable<unknown>): Promise<void> {
    // We need to consume the iterable but can't make a meaningful
    // binding for each element of the iteration.
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    for await (const _ of iterable) {
        // no-op body, just consume iterable
    }
}
