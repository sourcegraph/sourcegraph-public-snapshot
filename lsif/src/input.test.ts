import * as zlib from 'mz/zlib'
import { Edge, EdgeLabels, ElementTypes, Vertex, VertexLabels } from 'lsif-protocol'
import { elementValidator, processElements, processLsifInput, splitLines, validateLsifInput } from './input'
import { Readable, Writable } from 'stream'

describe('validateLsifInput', () => {
    it('should decode gzip', async () => {
        const lines = [
            '{"id": 1, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "foo.ts"}',
            '{"id": 2, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "bar.ts"}',
            '{"id": 3, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "baz.ts"}',
        ]

        const { output, written } = createWritable()
        const input = createReadable(lines.join('\n'))
        await validateLsifInput(input.pipe(zlib.createGzip()), output, elementValidator)
        expect((await zlib.gunzip(await written)).toString().trim()).toEqual(lines.join('\n'))
    })

    it('should fail without gzip', async () => {
        const lines = [
            '{"id": 1, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "foo.ts"}',
            '{"id": 2, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "bar.ts"}',
            '{"id": 3, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "baz.ts"}',
        ]

        const input = createReadable(lines.join('\n'))
        const { output } = createWritable()
        const promise = validateLsifInput(input, output, elementValidator)
        await expect(promise).rejects.toThrowError(new Error('incorrect header check'))
    })

    it('should wrap errors', async () => {
        const lines = [
            '{"id": 1, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "foo.ts"}',
            '{"id": 2, "type": "vertex", "label": "document"}',
            '{"id": 3, "type": "vertex", "label": "document", "languageId": "typescript", "uri": "baz.ts"}',
        ]

        const input = createReadable(lines.join('\n'))
        const { output } = createWritable()
        const promise = validateLsifInput(input.pipe(zlib.createGzip()), output, elementValidator)
        await expect(promise).rejects.toThrowError(
            new Error(
                'Failed to process line #2 ({"id": 2, "type": "vertex", "label": "document"}): should have required property \'data\''
            )
        )
    })
})

describe('processLsifInput', () => {
    it('should decode gzip', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const elements: (Vertex | Edge)[] = []
        const input = createReadable(lines.join('\n'))
        await processLsifInput(input.pipe(zlib.createGzip()), element => {
            elements.push(element)
        })

        expect(elements[0].type).toEqual(ElementTypes.vertex)
        expect(elements[0].label).toEqual(VertexLabels.project)
        expect(elements[1].type).toEqual(ElementTypes.vertex)
        expect(elements[1].label).toEqual(VertexLabels.document)
        expect(elements[2].type).toEqual(ElementTypes.edge)
        expect(elements[2].label).toEqual(EdgeLabels.item)
        expect(elements[3].type).toEqual(ElementTypes.edge)
        expect(elements[3].label).toEqual(EdgeLabels.moniker)
    })

    it('should fail without gzip', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const input = createReadable(lines.join('\n'))
        const promise = processLsifInput(input, () => {})
        await expect(promise).rejects.toThrowError(new Error('incorrect header check'))
    })

    it('should wrap errors', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const input = createReadable(lines.join('\n'))
        const promise = processLsifInput(input.pipe(zlib.createGzip()), element => {
            if (element.label === EdgeLabels.item) {
                throw new Error('foo')
            }
        })

        await expect(promise).rejects.toThrowError(
            new Error('Failed to process line #3 ({"type": "edge", "label": "item"}): foo')
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

describe('processElements', () => {
    it('should parse JSON', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const elements: (Vertex | Edge)[] = []
        await consume(
            processElements(generate(lines), element => {
                elements.push(element)
            })
        )

        expect(elements).toHaveLength(4)
        expect(elements[0].type).toEqual(ElementTypes.vertex)
        expect(elements[0].label).toEqual(VertexLabels.project)
        expect(elements[1].type).toEqual(ElementTypes.vertex)
        expect(elements[1].label).toEqual(VertexLabels.document)
        expect(elements[2].type).toEqual(ElementTypes.edge)
        expect(elements[2].label).toEqual(EdgeLabels.item)
        expect(elements[3].type).toEqual(ElementTypes.edge)
        expect(elements[3].label).toEqual(EdgeLabels.moniker)
    })

    it('should duplicate stream', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const duplicatedLines: string[] = []
        for await (const line of processElements(generate(lines), () => {})) {
            duplicatedLines.push(line)
        }

        expect(duplicatedLines).toEqual(lines)
    })

    it('should wrap errors', async () => {
        const lines = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge", "label": "moniker"}',
        ]

        const generator = processElements(generate(lines), element => {
            if (element.label === EdgeLabels.item) {
                throw new Error('foo')
            }
        })

        await expect(consume(generator)).rejects.toThrowError(
            new Error('Failed to process line #3 ({"type": "edge", "label": "item"}): foo')
        )
    })

    it('should wrap parse errors', async () => {
        const input = [
            '{"type": "vertex", "label": "project"}',
            '{"type": "vertex", "label": "document"}',
            '{"type": "edge", "label": "item"}',
            '{"type": "edge" "label": "moniker"}',
        ]

        await expect(consume(processElements(generate(input), () => {}))).rejects.toThrowError(
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

function createReadable(text: string): Readable {
    const input = new Readable()
    input.push(text)
    input.push(null)
    return input
}

function createWritable(): { output: Writable; written: Promise<Buffer> } {
    const buffers: Buffer[] = []
    const output = new Writable({
        write: (data, _, next) => {
            buffers.push(data)
            next()
        },
    })

    const written = new Promise<Buffer>((resolve, reject) => {
        output.on('error', reject).on('finish', () => resolve(Buffer.concat(buffers)))
    })

    return { output, written }
}

async function consume<T>(iterable: AsyncIterable<T>): Promise<void> {
    for await (const _ of iterable) {
        // no-op body, just consume iterable
    }
}
