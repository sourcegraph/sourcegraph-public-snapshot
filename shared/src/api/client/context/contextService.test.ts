import { createContextService } from './contextService'

describe('createContextService()', () => {
    describe('updateContext()', () => {
        it('merges properties', () => {
            const service = createContextService({ clientApplication: 'other' })
            service.updateContext({ a: 1, b: null, c: 2, d: 3, e: null })
            expect(service.data.value).toEqual({
                'clientApplication.isSourcegraph': false,
                'clientApplication.extensionAPIVersion.major': 3,
                a: 1,
                c: 2,
                d: 3,
            })
            service.updateContext({ a: null, b: 1, c: 3 })
            expect(service.data.value).toEqual({
                'clientApplication.isSourcegraph': false,
                'clientApplication.extensionAPIVersion.major': 3,
                b: 1,
                c: 3,
                d: 3,
            })
        })
    })
})
