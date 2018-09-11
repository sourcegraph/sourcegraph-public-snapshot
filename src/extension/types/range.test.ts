import assert from 'assert'
import * as sourcegraph from 'sourcegraph'
import { assertToJSON } from './common.test'
import { Position } from './position'
import { Range } from './range'

describe('Range', () => {
    it('constructs', () => {
        assert.throws(() => new Range(-1, 0, 0, 0))
        assert.throws(() => new Range(0, -1, 0, 0))
        assert.throws(() => new Range(new Position(0, 0), undefined as any))
        assert.throws(() => new Range(new Position(0, 0), null as any))
        assert.throws(() => new Range(undefined as any, new Position(0, 0)))
        assert.throws(() => new Range(null as any, new Position(0, 0)))

        const range = new Range(1, 0, 0, 0)
        assert.throws(() => {
            ;(range as any).start = null
        })
        assert.throws(() => {
            ;(range as any).start = new Position(0, 3)
        })
    })

    it('toJSON', () => {
        const range = new Range(1, 2, 3, 4)
        assertToJSON(range, { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } })
    })

    it('sorting', () => {
        // sorts start/end
        let range = new Range(1, 0, 0, 0)
        assert.strictEqual(range.start.line, 0)
        assert.strictEqual(range.end.line, 1)

        range = new Range(0, 0, 1, 0)
        assert.strictEqual(range.start.line, 0)
        assert.strictEqual(range.end.line, 1)
    })

    it('isEmpty|isSingleLine', () => {
        let range = new Range(1, 0, 0, 0)
        assert.ok(!range.isEmpty)
        assert.ok(!range.isSingleLine)

        range = new Range(1, 1, 1, 1)
        assert.ok(range.isEmpty)
        assert.ok(range.isSingleLine)

        range = new Range(0, 1, 0, 11)
        assert.ok(!range.isEmpty)
        assert.ok(range.isSingleLine)

        range = new Range(0, 0, 1, 1)
        assert.ok(!range.isEmpty)
        assert.ok(!range.isSingleLine)
    })

    it('contains', () => {
        const range = new Range(1, 1, 2, 11)

        assert.ok(range.contains(range.start))
        assert.ok(range.contains(range.end))
        assert.ok(range.contains(range))

        assert.ok(!range.contains(new Range(1, 0, 2, 11)))
        assert.ok(!range.contains(new Range(0, 1, 2, 11)))
        assert.ok(!range.contains(new Range(1, 1, 2, 12)))
        assert.ok(!range.contains(new Range(1, 1, 3, 11)))
    })

    it('intersection', () => {
        const range = new Range(1, 1, 2, 11)
        let res: sourcegraph.Range | undefined

        res = range.intersection(range)
        assert.strictEqual(res && res.start.line, 1)
        assert.strictEqual(res && res.start.character, 1)
        assert.strictEqual(res && res.end.line, 2)
        assert.strictEqual(res && res.end.character, 11)

        res = range.intersection(new Range(2, 12, 4, 0))
        assert.strictEqual(res, undefined)

        res = range.intersection(new Range(0, 0, 1, 0))
        assert.strictEqual(res, undefined)

        res = range.intersection(new Range(0, 0, 1, 1))
        assert.ok(res && res.isEmpty)
        assert.strictEqual(res && res.start.line, 1)
        assert.strictEqual(res && res.start.character, 1)

        res = range.intersection(new Range(2, 11, 61, 1))
        assert.ok(res && res.isEmpty)
        assert.strictEqual(res && res.start.line, 2)
        assert.strictEqual(res && res.start.character, 11)

        assert.throws(() => range.intersection(null as any))
        assert.throws(() => range.intersection(undefined as any))
    })

    it('union', () => {
        let ran1 = new Range(0, 0, 5, 5)
        assert.ok(ran1.union(new Range(0, 0, 1, 1)) === ran1)

        let res: sourcegraph.Range
        res = ran1.union(new Range(2, 2, 9, 9))
        assert.ok(res.start === ran1.start)
        assert.strictEqual(res.end.line, 9)
        assert.strictEqual(res.end.character, 9)

        ran1 = new Range(2, 1, 5, 3)
        res = ran1.union(new Range(1, 0, 4, 2))
        assert.ok(res.end === ran1.end)
        assert.strictEqual(res.start.line, 1)
        assert.strictEqual(res.start.character, 0)
    })

    it('with', () => {
        const range = new Range(1, 1, 2, 11)

        assert.ok(range.with(range.start) === range)
        assert.ok(range.with(undefined, range.end) === range)
        assert.ok(range.with(range.start, range.end) === range)
        assert.ok(range.with(new Position(1, 1)) === range)
        assert.ok(range.with(undefined, new Position(2, 11)) === range)
        assert.ok(range.with() === range)
        assert.ok(range.with({ start: range.start }) === range)
        assert.ok(range.with({ start: new Position(1, 1) }) === range)
        assert.ok(range.with({ end: range.end }) === range)
        assert.ok(range.with({ end: new Position(2, 11) }) === range)

        let res = range.with(undefined, new Position(9, 8))
        assert.strictEqual(res.end.line, 9)
        assert.strictEqual(res.end.character, 8)
        assert.strictEqual(res.start.line, 1)
        assert.strictEqual(res.start.character, 1)

        res = range.with({ end: new Position(9, 8) })
        assert.strictEqual(res.end.line, 9)
        assert.strictEqual(res.end.character, 8)
        assert.strictEqual(res.start.line, 1)
        assert.strictEqual(res.start.character, 1)

        res = range.with({ end: new Position(9, 8), start: new Position(2, 3) })
        assert.strictEqual(res.end.line, 9)
        assert.strictEqual(res.end.character, 8)
        assert.strictEqual(res.start.line, 2)
        assert.strictEqual(res.start.character, 3)

        assert.throws(() => range.with(null as any))
        assert.throws(() => range.with(undefined, null as any))
    })
})
