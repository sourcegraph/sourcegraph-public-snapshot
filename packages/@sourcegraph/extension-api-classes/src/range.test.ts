import * as sourcegraph from 'sourcegraph'
import { Position } from './position'
import { Range } from './range'
import { assertToJSON } from './testHelpers'

describe('Range', () => {
    test('constructs', () => {
        expect(() => new Range(-1, 0, 0, 0)).toThrow()
        expect(() => new Range(0, -1, 0, 0)).toThrow()
        expect(() => new Range(new Position(0, 0), undefined as any)).toThrow()
        expect(() => new Range(new Position(0, 0), null as any)).toThrow()
        expect(() => new Range(undefined as any, new Position(0, 0))).toThrow()
        expect(() => new Range(null as any, new Position(0, 0))).toThrow()

        const range = new Range(1, 0, 0, 0)
        expect(() => {
            ;(range as any).start = null
        }).toThrow()
        expect(() => {
            ;(range as any).start = new Position(0, 3)
        }).toThrow()
    })

    test('toJSON', () => {
        const range = new Range(1, 2, 3, 4)
        assertToJSON(range, { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } })
    })

    test('toPlain', () => {
        const range = new Range(1, 2, 3, 4)
        expect(range.toPlain()).toEqual({ start: { line: 1, character: 2 }, end: { line: 3, character: 4 } })
    })

    test('sorting', () => {
        // sorts start/end
        let range = new Range(1, 0, 0, 0)
        expect(range.start.line).toBe(0)
        expect(range.end.line).toBe(1)

        range = new Range(0, 0, 1, 0)
        expect(range.start.line).toBe(0)
        expect(range.end.line).toBe(1)
    })

    test('isEmpty|isSingleLine', () => {
        let range = new Range(1, 0, 0, 0)
        expect(!range.isEmpty).toBeTruthy()
        expect(!range.isSingleLine).toBeTruthy()

        range = new Range(1, 1, 1, 1)
        expect(range.isEmpty).toBeTruthy()
        expect(range.isSingleLine).toBeTruthy()

        range = new Range(0, 1, 0, 11)
        expect(!range.isEmpty).toBeTruthy()
        expect(range.isSingleLine).toBeTruthy()

        range = new Range(0, 0, 1, 1)
        expect(!range.isEmpty).toBeTruthy()
        expect(!range.isSingleLine).toBeTruthy()
    })

    test('contains', () => {
        const range = new Range(1, 1, 2, 11)

        expect(range.contains(range.start)).toBeTruthy()
        expect(range.contains(range.end)).toBeTruthy()
        expect(range.contains(range)).toBeTruthy()

        expect(!range.contains(new Range(1, 0, 2, 11))).toBeTruthy()
        expect(!range.contains(new Range(0, 1, 2, 11))).toBeTruthy()
        expect(!range.contains(new Range(1, 1, 2, 12))).toBeTruthy()
        expect(!range.contains(new Range(1, 1, 3, 11))).toBeTruthy()
    })

    test('intersection', () => {
        const range = new Range(1, 1, 2, 11)
        let res: sourcegraph.Range | undefined

        res = range.intersection(range)
        expect(res && res.start.line).toBe(1)
        expect(res && res.start.character).toBe(1)
        expect(res && res.end.line).toBe(2)
        expect(res && res.end.character).toBe(11)

        res = range.intersection(new Range(2, 12, 4, 0))
        expect(res).toBe(undefined)

        res = range.intersection(new Range(0, 0, 1, 0))
        expect(res).toBe(undefined)

        res = range.intersection(new Range(0, 0, 1, 1))
        expect(res && res.isEmpty).toBeTruthy()
        expect(res && res.start.line).toBe(1)
        expect(res && res.start.character).toBe(1)

        res = range.intersection(new Range(2, 11, 61, 1))
        expect(res && res.isEmpty).toBeTruthy()
        expect(res && res.start.line).toBe(2)
        expect(res && res.start.character).toBe(11)

        expect(() => range.intersection(null as any)).toThrow()
        expect(() => range.intersection(undefined as any)).toThrow()
    })

    test('union', () => {
        let ran1 = new Range(0, 0, 5, 5)
        expect(ran1.union(new Range(0, 0, 1, 1)) === ran1).toBeTruthy()

        let res: sourcegraph.Range
        res = ran1.union(new Range(2, 2, 9, 9))
        expect(res.start === ran1.start).toBeTruthy()
        expect(res.end.line).toBe(9)
        expect(res.end.character).toBe(9)

        ran1 = new Range(2, 1, 5, 3)
        res = ran1.union(new Range(1, 0, 4, 2))
        expect(res.end === ran1.end).toBeTruthy()
        expect(res.start.line).toBe(1)
        expect(res.start.character).toBe(0)
    })

    test('with', () => {
        const range = new Range(1, 1, 2, 11)

        expect(range.with(range.start) === range).toBeTruthy()
        expect(range.with(undefined, range.end) === range).toBeTruthy()
        expect(range.with(range.start, range.end) === range).toBeTruthy()
        expect(range.with(new Position(1, 1)) === range).toBeTruthy()
        expect(range.with(undefined, new Position(2, 11)) === range).toBeTruthy()
        expect(range.with() === range).toBeTruthy()
        expect(range.with({ start: range.start }) === range).toBeTruthy()
        expect(range.with({ start: new Position(1, 1) }) === range).toBeTruthy()
        expect(range.with({ end: range.end }) === range).toBeTruthy()
        expect(range.with({ end: new Position(2, 11) }) === range).toBeTruthy()

        let res = range.with(undefined, new Position(9, 8))
        expect(res.end.line).toBe(9)
        expect(res.end.character).toBe(8)
        expect(res.start.line).toBe(1)
        expect(res.start.character).toBe(1)

        res = range.with({ end: new Position(9, 8) })
        expect(res.end.line).toBe(9)
        expect(res.end.character).toBe(8)
        expect(res.start.line).toBe(1)
        expect(res.start.character).toBe(1)

        res = range.with({ end: new Position(9, 8), start: new Position(2, 3) })
        expect(res.end.line).toBe(9)
        expect(res.end.character).toBe(8)
        expect(res.start.line).toBe(2)
        expect(res.start.character).toBe(3)

        expect(() => range.with(null as any)).toThrow()
        expect(() => range.with(undefined, null as any)).toThrow()
    })
})
