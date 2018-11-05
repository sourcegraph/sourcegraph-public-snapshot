import * as assert from 'assert'
import { LinkedMap, Touch } from './linkedMap'

describe('LinkedMap', () => {
    it('simple', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        assert.deepStrictEqual(Array.from(map.keys()), ['ak', 'bk'])
        assert.deepStrictEqual(Array.from(map.values()), ['av', 'bv'])
    })

    it('touch first', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('ak', 'av', Touch.First)
        assert.deepStrictEqual(Array.from(map.keys()), ['ak'])
        assert.deepStrictEqual(Array.from(map.values()), ['av'])
    })

    it('touch last', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('ak', 'av', Touch.Last)
        assert.deepStrictEqual(Array.from(map.keys()), ['ak'])
        assert.deepStrictEqual(Array.from(map.values()), ['av'])
    })

    it('touch first 2', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('bk', 'bv', Touch.First)
        assert.deepStrictEqual(Array.from(map.keys()), ['bk', 'ak'])
        assert.deepStrictEqual(Array.from(map.values()), ['bv', 'av'])
    })

    it('touch last 2', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ak', 'av', Touch.Last)
        assert.deepStrictEqual(Array.from(map.keys()), ['bk', 'ak'])
        assert.deepStrictEqual(Array.from(map.values()), ['bv', 'av'])
    })

    it('touch first from middle', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ck', 'cv')
        map.set('bk', 'bv', Touch.First)
        assert.deepStrictEqual(Array.from(map.keys()), ['bk', 'ak', 'ck'])
        assert.deepStrictEqual(Array.from(map.values()), ['bv', 'av', 'cv'])
    })

    it('touch last from middle', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ck', 'cv')
        map.set('bk', 'bv', Touch.Last)
        assert.deepStrictEqual(Array.from(map.keys()), ['ak', 'ck', 'bk'])
        assert.deepStrictEqual(Array.from(map.values()), ['av', 'cv', 'bv'])
    })
})
