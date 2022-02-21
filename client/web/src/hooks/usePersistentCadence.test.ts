import { usePersistentCadence, reset } from './usePersistentCadence'

const testKey = 'test'
const testKey2 = 'test2'
const testCadence = 5

beforeEach(() => {
    localStorage.clear()
    reset()
})

describe('usePersistentCadence', () => {
    it('returns true on first run', () => {
        expect(usePersistentCadence(testKey, testCadence)).toBe(true)
    })

    it('returns true on second run in the same session', () => {
        usePersistentCadence(testKey, testCadence)

        expect(usePersistentCadence(testKey, testCadence)).toBe(true)
    })

    it('returns true on Nth run', () => {
        localStorage.setItem(testKey, testCadence.toString())

        expect(usePersistentCadence(testKey, testCadence)).toBe(true)
    })

    it('returns true on 2Nth run', () => {
        localStorage.setItem(testKey, (2 * testCadence).toString())

        expect(usePersistentCadence(testKey, testCadence)).toBe(true)
    })

    it('returns false on (2N+1)th run', () => {
        localStorage.setItem(testKey, (2 * testCadence + 1).toString())

        expect(usePersistentCadence(testKey, testCadence)).toBe(false)
    })

    it('handles different keys separately', () => {
        localStorage.setItem(testKey, (2 * testCadence + 1).toString())

        expect(usePersistentCadence(testKey2, testCadence)).toBe(true)
    })

    it('returns true on {N+shift}th run', () => {
        const testShift = 2
        localStorage.setItem(testKey, (testCadence + testShift).toString())

        expect(usePersistentCadence(testKey, testCadence, testShift)).toBe(true)
    })
})
