import * as zlib from 'mz/zlib'
import { parseJsonLines, readGzippedJsonElements, splitLines, validateLsifElements } from './input'
import { Readable } from 'stream'

describe('readGzippedJsonElements', () => {
    it('should decode gzip', async () => {
        const lines = [
            { type: 'vertex', label: 'project' },
            { type: 'vertex', label: 'document' },
            { type: 'edge', label: 'item' },
            { type: 'edge', label: 'moniker' },
        ]

        const elements: unknown[] = []
        const input = Readable.from(lines.map(l => JSON.stringify(l)).join('\n'))
        for await (const element of readGzippedJsonElements(input.pipe(zlib.createGzip()))) {
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

        const input = Readable.from(lines.join('\n'))
        await expect(consume(readGzippedJsonElements(input))).rejects.toThrowError(new Error('incorrect header check'))
    })
})

describe('validateLsifElements', () => {
    it('should reject invalid LSIF', async () => {
        const lines = [
            { id: 1, type: 'vertex', label: 'whatisthis', languageId: 'typescript', uri: 'foo.ts' },
            { id: 2, type: 'vertex', label: 'document' },
            { id: 3, type: 'vertex', label: 'document', languageId: 'typescript', uri: 'baz.ts' },
        ]

        const input = generate(lines)
        const promise = consume(validateLsifElements(input))
        await expect(promise).rejects.toThrowError(
            new Error(
                'Invalid LSIF element at index #1 ({"id":1,"type":"vertex","label":"whatisthis","languageId":"typescript","uri":"foo.ts"}): should NOT have additional properties'
            )
        )
    })
})

describe('splitLines', () => {
    it('should split input by newline', async () => {
        const chunks = ['foo\n', 'bar', '\nbaz\n\nbonk\nqu', 'ux']

        const lines = []
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

        const elements: any[] = []
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
    for await (const _ of iterable) {
        // no-op body, just consume iterable
    }
}
