import { sourcegraphPlugin } from './plugin'

describe('sourcegraph', () => {
    it('should export plugin', () => {
        expect(sourcegraphPlugin).toBeDefined()
    })
})
