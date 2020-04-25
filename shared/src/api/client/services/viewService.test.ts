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

    test('get  nonexistent view', () => {
        const viewService = createViewService()
        scheduler().run(({ expectObservable }) => {
            expectObservable(viewService.get('v', {})).toBe('a', {
                a: null,
            })
        })
    })

    test('register then get', () => {
        const viewService = createViewService()
        scheduler().run(({ cold, expectObservable }) => {
            viewService.register('v', () =>
                cold<View>('bc', { b: { title: 'b', content: [] }, c: { title: 'c', content: [] } })
            )
            expectObservable(viewService.get('v', {})).toBe('bc', {
                b: { title: 'b', content: [] },
                c: { title: 'c', content: [] },
            })
        })
    })

    test('register, unsubscribe, then get', () => {
        const viewService = createViewService()
        scheduler().run(({ cold, expectObservable }) => {
            const subscription = viewService.register('v', () =>
                cold<View>('bc', { b: { title: 'b', content: [] }, c: { title: 'c', content: [] } })
            )
            subscription.unsubscribe()
            expectObservable(viewService.get('v', {})).toBe('d', {
                d: null,
            })
        })
    })
})
