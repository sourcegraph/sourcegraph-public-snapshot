import { shave } from './strings'

describe('utils/string', () => {
    describe('shave()', () => {
        it('turns trailing whitespace into a single whitespace', () => {
            expect(shave('    a    b   c   d   ')).toBe(' a b c d ')
        })
    })
})
