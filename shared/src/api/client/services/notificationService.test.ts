import { throwError, Unsubscribable, of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import * as sourcegraph from 'sourcegraph'
import { createNotificationService } from './notificationService'
import { NotificationType } from '@sourcegraph/extension-api-classes'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const NOTIFICATION_1: sourcegraph.Notification = { title: 't1', type: NotificationType.Info }

const NOTIFICATION_2: sourcegraph.Notification = { title: 't2', type: NotificationType.Error }

const SCOPE: sourcegraph.NotificationScope = 'global' as sourcegraph.NotificationScope.Global

describe('NotificationService', () => {
    test('no providers yields empty array', () =>
        scheduler().run(({ expectObservable }) =>
            expectObservable(createNotificationService(false).observeNotifications(SCOPE)).toBe('a', {
                a: [],
            })
        ))

    test('single provider', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createNotificationService(false)
            service.registerNotificationProvider('', {
                provideNotifications: () =>
                    cold<sourcegraph.Notification[] | null>('abcd', {
                        a: null,
                        b: [NOTIFICATION_1],
                        c: [NOTIFICATION_1, NOTIFICATION_2],
                        d: null,
                    }),
            })
            expectObservable(service.observeNotifications(SCOPE)).toBe('abcd', {
                a: [],
                b: [NOTIFICATION_1],
                c: [NOTIFICATION_1, NOTIFICATION_2],
                d: [],
            })
        })
    })

    test('merges results from multiple providers', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createNotificationService(false)
            const unsub1 = service.registerNotificationProvider('1', {
                provideNotifications: () => of([NOTIFICATION_1]),
            })
            let unsub2: Unsubscribable
            cold('-bc', {
                b: () => {
                    unsub2 = service.registerNotificationProvider('2', {
                        provideNotifications: () => of([NOTIFICATION_2]),
                    })
                },
                c: () => {
                    unsub1.unsubscribe()
                    unsub2.unsubscribe()
                },
            }).subscribe(f => f())
            expectObservable(service.observeNotifications(SCOPE)).toBe('ab(cd)', {
                a: [NOTIFICATION_1],
                b: [NOTIFICATION_1, NOTIFICATION_2],
                c: [NOTIFICATION_2],
                d: [],
            })
        })
    })

    test('suppresses errors', () => {
        scheduler().run(({ expectObservable }) => {
            const service = createNotificationService(false)
            service.registerNotificationProvider('a', {
                provideNotifications: () => throwError(new Error('x')),
            })
            expectObservable(service.observeNotifications(SCOPE)).toBe('a', {
                a: [],
            })
        })
    })

    test('enforces unique registration types', () => {
        const service = createNotificationService(false)
        service.registerNotificationProvider('a', {
            provideNotifications: () => [],
        })
        expect(() =>
            service.registerNotificationProvider('a', {
                provideNotifications: () => [],
            })
        ).toThrowError(/already registered/)
    })
})
