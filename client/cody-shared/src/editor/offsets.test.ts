import { DocumentOffsets } from './offsets'

describe('DocumentOffsets', () => {
    it('converts offsets to positions and back and back properly', () => {
        const texts = [
            `ABC
DEF
GHI`,
            `ABC
DEF
GHI
`,
        ]
        for (const text of texts) {
            const offset = new DocumentOffsets(text)

            for (let i = 0; i < text.length + 1; i++) {
                const pos = offset.position(i)
                const o2 = offset.offset(pos)
                expect(i).toEqual(o2)
                expect(pos).toEqual(offset.position(o2))
            }
        }
    })

    it('provides the right line range', () => {
        const text = `Hello
World
More`
        const offset = new DocumentOffsets(text)

        expect('Hello\n').toEqual(offset.getLine(0))
        expect('World\n').toEqual(offset.getLine(1))
        expect('More').toEqual(offset.getLine(2))
    })
})
