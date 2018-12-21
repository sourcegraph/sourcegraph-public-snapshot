import { LinkedMap, Touch } from './linkedMap'

describe('LinkedMap', () => {
    test('simple', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        expect(Array.from(map.keys())).toEqual(['ak', 'bk'])
        expect(Array.from(map.values())).toEqual(['av', 'bv'])
    })

    test('touch first', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('ak', 'av', Touch.First)
        expect(Array.from(map.keys())).toEqual(['ak'])
        expect(Array.from(map.values())).toEqual(['av'])
    })

    test('touch last', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('ak', 'av', Touch.Last)
        expect(Array.from(map.keys())).toEqual(['ak'])
        expect(Array.from(map.values())).toEqual(['av'])
    })

    test('touch first 2', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('bk', 'bv', Touch.First)
        expect(Array.from(map.keys())).toEqual(['bk', 'ak'])
        expect(Array.from(map.values())).toEqual(['bv', 'av'])
    })

    test('touch last 2', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ak', 'av', Touch.Last)
        expect(Array.from(map.keys())).toEqual(['bk', 'ak'])
        expect(Array.from(map.values())).toEqual(['bv', 'av'])
    })

    test('touch first from middle', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ck', 'cv')
        map.set('bk', 'bv', Touch.First)
        expect(Array.from(map.keys())).toEqual(['bk', 'ak', 'ck'])
        expect(Array.from(map.values())).toEqual(['bv', 'av', 'cv'])
    })

    test('touch last from middle', () => {
        const map = new LinkedMap<string, string>()
        map.set('ak', 'av')
        map.set('bk', 'bv')
        map.set('ck', 'cv')
        map.set('bk', 'bv', Touch.Last)
        expect(Array.from(map.keys())).toEqual(['ak', 'ck', 'bk'])
        expect(Array.from(map.values())).toEqual(['av', 'cv', 'bv'])
    })
})
