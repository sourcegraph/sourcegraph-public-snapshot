import { NEVER } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { createViewService, View } from './viewService'

const scheduler = (): TestScheduler => new TestScheduler((actual, expected) => expect(actual).toEqual(expected))

describe('ViewService', () => {
    test('throws on ID conflict', () => {
        const viewService = createViewService()
        viewService.register('v', () => NEVER)
        expect(() => viewService.register('v', () => NEVER)).toThrow()
    })

    test('get', () => {
        const viewService = createViewService()
        scheduler().run(({ cold, expectObservable }) => {
            viewService.register('v', () =>
                cold<View>('ab', { a: { title: 'a', content: [] }, b: { title: 'b', content: [] } })
            )
            expectObservable(viewService.get('v', {})).toBe('ab', {
                a: { title: 'a', content: [] },
                b: { title: 'b', content: [] },
            })
        })
    })
})
