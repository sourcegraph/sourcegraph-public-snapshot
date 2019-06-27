import { throwError, Unsubscribable, of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import * as sourcegraph from 'sourcegraph'
import { createChecklistService } from './checklistService'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const CHECKLIST_ITEM_1: sourcegraph.ChecklistItem = { title: 't1' }

const CHECKLIST_ITEM_2: sourcegraph.ChecklistItem = { title: 't2' }

const SCOPE: sourcegraph.ChecklistScope = 'global' as sourcegraph.ChecklistScope.Global

describe('ChecklistService', () => {
    test('no providers yields empty array', () =>
        scheduler().run(({ expectObservable }) =>
            expectObservable(createChecklistService(false).observeChecklistItems(SCOPE)).toBe('a', {
                a: [],
            })
        ))

    test('single provider', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createChecklistService(false)
            service.registerChecklistProvider('', {
                provideChecklistItems: () =>
                    cold<sourcegraph.ChecklistItem[] | null>('abcd', {
                        a: null,
                        b: [CHECKLIST_ITEM_1],
                        c: [CHECKLIST_ITEM_1, CHECKLIST_ITEM_2],
                        d: null,
                    }),
            })
            expectObservable(service.observeChecklistItems(SCOPE)).toBe('abcd', {
                a: [],
                b: [CHECKLIST_ITEM_1],
                c: [CHECKLIST_ITEM_1, CHECKLIST_ITEM_2],
                d: [],
            })
        })
    })

    test('merges results from multiple providers', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createChecklistService(false)
            const unsub1 = service.registerChecklistProvider('1', {
                provideChecklistItems: () => of([CHECKLIST_ITEM_1]),
            })
            let unsub2: Unsubscribable
            cold('-bc', {
                b: () => {
                    unsub2 = service.registerChecklistProvider('2', {
                        provideChecklistItems: () => of([CHECKLIST_ITEM_2]),
                    })
                },
                c: () => {
                    unsub1.unsubscribe()
                    unsub2.unsubscribe()
                },
            }).subscribe(f => f())
            expectObservable(service.observeChecklistItems(SCOPE)).toBe('ab(cd)', {
                a: [CHECKLIST_ITEM_1],
                b: [CHECKLIST_ITEM_1, CHECKLIST_ITEM_2],
                c: [CHECKLIST_ITEM_2],
                d: [],
            })
        })
    })

    test('suppresses errors', () => {
        scheduler().run(({ expectObservable }) => {
            const service = createChecklistService(false)
            service.registerChecklistProvider('a', {
                provideChecklistItems: () => throwError(new Error('x')),
            })
            expectObservable(service.observeChecklistItems(SCOPE)).toBe('a', {
                a: [],
            })
        })
    })

    test('enforces unique registration types', () => {
        const service = createChecklistService(false)
        service.registerChecklistProvider('a', {
            provideChecklistItems: () => [],
        })
        expect(() =>
            service.registerChecklistProvider('a', {
                provideChecklistItems: () => [],
            })
        ).toThrowError(/already registered/)
    })
})
