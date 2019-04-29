import { Position } from './position'
import { assertToJSON } from './testHelpers'

describe('Position', () => {
    test('constructs', () => {
        expect(() => new Position(-1, 0)).toThrow()
        expect(() => new Position(0, -1)).toThrow()

        const pos = new Position(0, 0)
        expect(() => ((pos as any).line = -1)).toThrow()
        expect(() => ((pos as any).character = -1)).toThrow()
        expect(() => ((pos as any).line = 12)).toThrow()

        const { line, character } = pos.toJSON()
        expect(line).toBe(0)
        expect(character).toBe(0)
    })

    test('toJSON', () => {
        const pos = new Position(4, 2)
        assertToJSON(pos, { line: 4, character: 2 })
    })

    test('isBefore(OrEqual)?', () => {
        const p1 = new Position(1, 3)
        const p2 = new Position(1, 2)
        const p3 = new Position(0, 4)

        expect(p1.isBeforeOrEqual(p1)).toBeTruthy()
        expect(!p1.isBefore(p1)).toBeTruthy()
        expect(p2.isBefore(p1)).toBeTruthy()
        expect(p3.isBefore(p2)).toBeTruthy()
    })

    test('isAfter(OrEqual)?', () => {
        const p1 = new Position(1, 3)
        const p2 = new Position(1, 2)
        const p3 = new Position(0, 4)

        expect(p1.isAfterOrEqual(p1)).toBeTruthy()
        expect(!p1.isAfter(p1)).toBeTruthy()
        expect(p1.isAfter(p2)).toBeTruthy()
        expect(p2.isAfter(p3)).toBeTruthy()
        expect(p1.isAfter(p3)).toBeTruthy()
    })

    test('compareTo', () => {
        const p1 = new Position(1, 3)
        const p2 = new Position(1, 2)
        const p3 = new Position(0, 4)

        expect(p1.compareTo(p1)).toBe(0)
        expect(p2.compareTo(p1)).toBe(-1)
        expect(p1.compareTo(p2)).toBe(1)
        expect(p2.compareTo(p3)).toBe(1)
        expect(p1.compareTo(p3)).toBe(1)
    })

    test('translate', () => {
        const p1 = new Position(1, 3)

        expect(p1.translate() === p1).toBeTruthy()
        expect(p1.translate({}) === p1).toBeTruthy()
        expect(p1.translate(0, 0) === p1).toBeTruthy()
        expect(p1.translate(0) === p1).toBeTruthy()
        expect(p1.translate(undefined, 0) === p1).toBeTruthy()
        expect(p1.translate(undefined) === p1).toBeTruthy()

        let res = p1.translate(-1)
        expect(res.line).toBe(0)
        expect(res.character).toBe(3)

        res = p1.translate({ lineDelta: -1 })
        expect(res.line).toBe(0)
        expect(res.character).toBe(3)

        res = p1.translate(undefined, -1)
        expect(res.line).toBe(1)
        expect(res.character).toBe(2)

        res = p1.translate({ characterDelta: -1 })
        expect(res.line).toBe(1)
        expect(res.character).toBe(2)

        res = p1.translate(11)
        expect(res.line).toBe(12)
        expect(res.character).toBe(3)

        expect(() => p1.translate(null as any)).toThrow()
        expect(() => p1.translate(null as any, null as any)).toThrow()
        expect(() => p1.translate(-2)).toThrow()
        expect(() => p1.translate({ lineDelta: -2 })).toThrow()
        expect(() => p1.translate(-2, null as any)).toThrow()
        expect(() => p1.translate(0, -4)).toThrow()
    })

    test('with', () => {
        const p1 = new Position(1, 3)

        expect(p1.with() === p1).toBeTruthy()
        expect(p1.with(1) === p1).toBeTruthy()
        expect(p1.with(undefined, 3) === p1).toBeTruthy()
        expect(p1.with(1, 3) === p1).toBeTruthy()
        expect(p1.with(undefined) === p1).toBeTruthy()
        expect(p1.with({ line: 1 }) === p1).toBeTruthy()
        expect(p1.with({ character: 3 }) === p1).toBeTruthy()
        expect(p1.with({ line: 1, character: 3 }) === p1).toBeTruthy()

        const p2 = p1.with({ line: 0, character: 11 })
        expect(p2.line).toBe(0)
        expect(p2.character).toBe(11)

        expect(() => p1.with(null as any)).toThrow()
        expect(() => p1.with(-9)).toThrow()
        expect(() => p1.with(0, -9)).toThrow()
        expect(() => p1.with({ line: -1 })).toThrow()
        expect(() => p1.with({ character: -1 })).toThrow()
    })
})
