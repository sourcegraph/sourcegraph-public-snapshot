import { createBatcher } from './batch'
import { range } from 'lodash'

describe('createBatcher', () => {
    it('should traverse entire tree', () => {
        const values = gatherValues('foo', [
            ...range(10).map(i => `bar/${i}.ts`),
            ...range(10).map(i => `bar/baz/${i}.ts`),
            ...range(10).map(i => `bar/baz/bonk/${i}.ts`),
        ])

        expect(values).toEqual([[''], ['foo'], ['foo/bar'], ['foo/bar/baz'], ['foo/bar/baz/bonk']])
    })

    it('should batch entries at same depth', () => {
        const values = gatherValues(
            'foo',
            ['bar', 'baz', 'bonk'].map(d => `${d}/sub/file.ts`)
        )

        expect(values).toEqual([
            [''],
            ['foo'],
            ['foo/bar', 'foo/baz', 'foo/bonk'],
            ['foo/bar/sub', 'foo/baz/sub', 'foo/bonk/sub'],
        ])
    })

    it('should batch entries at same depth (wide)', () => {
        const is = range(0, 5)
        const ds = ['bar', 'baz', 'bonk']

        const values = gatherValues(
            'foo',
            ds.flatMap(d => is.map(i => `${d}/${i}/file.ts`))
        )

        expect(values).toEqual([
            [''],
            ['foo'],
            ['foo/bar', 'foo/baz', 'foo/bonk'],
            [
                'foo/bar/0',
                'foo/bar/1',
                'foo/bar/2',
                'foo/bar/3',
                'foo/bar/4',
                'foo/baz/0',
                'foo/baz/1',
                'foo/baz/2',
                'foo/baz/3',
                'foo/baz/4',
                'foo/bonk/0',
                'foo/bonk/1',
                'foo/bonk/2',
                'foo/bonk/3',
                'foo/bonk/4',
            ],
        ])
    })

    it('should cut subtrees that do not exist', () => {
        const ds = ['bar', 'baz', 'bonk']
        const ss = ['a', 'b', 'c']
        const is = range(1, 4)

        const blacklist = ['foo/bar', 'foo/baz/a', 'foo/bonk/a/1', 'foo/bonk/a/2', 'foo/bonk/b/1', 'foo/bonk/b/3']

        const values = gatherValues(
            'foo',
            ds.flatMap(d => ss.flatMap(s => is.map(i => `${d}/${s}/${i}/sub/file.ts`))),
            blacklist
        )

        const prune = (paths: string[]): string[] =>
            // filter out all proper descendants of the blacklist
            paths.filter(p => blacklist.includes(p) || !blacklist.some(b => p.includes(b)))

        expect(values).toEqual([
            [''],
            ['foo'],
            prune(ds.map(d => `foo/${d}`)),
            prune(ds.flatMap(d => ss.map(s => `foo/${d}/${s}`))),
            prune(ds.flatMap(d => ss.flatMap(s => is.map(i => `foo/${d}/${s}/${i}`)))),
            prune(ds.flatMap(d => ss.flatMap(s => is.map(i => `foo/${d}/${s}/${i}/sub`)))),
        ])
    })
})

function gatherValues(root: string, documentPaths: string[], blacklist: string[] = []): string[][] {
    const batcher = createBatcher(root, documentPaths)

    let done: boolean | undefined
    let batch: string[] | void | undefined

    const all = []
    while (true) {
        ;({ value: batch, done } = batcher.next((batch || []).filter(x => !blacklist?.includes(x))))
        if (done || !batch) {
            break
        }

        all.push(batch)
    }

    return all
}
