import { toPartialUtf8String } from './utils'

describe('toPartialUtf8String', () => {
    it('should decode single-byte characters', () => {
        const { str, buf } = toPartialUtf8String(Buffer.from('hello, world', 'utf-8'))
        expect(str).toBe('hello, world')
        expect(buf.length).toBe(0)
    })
    it('should decode multibyte characters', () => {
        const { str, buf } = toPartialUtf8String(Buffer.from('今日は、世界', 'utf-8'))
        expect(str).toBe('今日は、世界')
        expect(buf.length).toBe(0)
    })
    it('should split if the last byte is the initial byte of a multibyte character', () => {
        const { str, buf } = toPartialUtf8String(Buffer.from([0x48, 0x69, 0x20, 0xef]))
        expect(str).toBe('Hi ')
        expect(buf).toEqual(Buffer.from([0xef]))
    })
    it('should split if the trailing bytes are the start of a multibyte character', () => {
        const { str, buf } = toPartialUtf8String(Buffer.from([0x59, 0x6f, 0x21, 0xf0, 0x8a, 0x8b]))
        expect(str).toBe('Yo!')
        expect(buf).toEqual(Buffer.from([0xf0, 0x8a, 0x8b]))
    })
})
