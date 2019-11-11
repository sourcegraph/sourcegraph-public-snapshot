import { from, Observable, of, Subject, Subscription } from 'rxjs'
import { bufferCount, map, switchMap, toArray } from 'rxjs/operators'
import * as sinon from 'sinon'
import { createBarrier } from '../../../../shared/src/api/integration-test/testHelpers'
import { MutationRecordLike } from '../../shared/util/dom'
import { trackViews, ViewResolver } from './views'

const FIXTURE_HTML = `
    <div id="parent">
        <div class="view" id="1"></div>
        <div class="view" id="2"></div>
        <div class="view" id="3"></div>
    </div>
`

describe('trackViews()', () => {
    let subscriptions = new Subscription()

    beforeEach(() => {
        document.body.innerHTML = FIXTURE_HTML
    })

    afterAll(() => {
        subscriptions.unsubscribe()
        subscriptions = new Subscription()
        document.body.innerHTML = ''
    })

    test('detects all views on the page', async () => {
        const mutations: Observable<MutationRecordLike[]> = of([{ addedNodes: [document.body], removedNodes: [] }])
        const views = await mutations
            .pipe(trackViews([{ selector: '.view', resolveView: element => ({ element }) }]), toArray())
            .toPromise()
        expect(views.map(({ element }) => element.id)).toEqual(['1', '2', '3'])
    })

    test('detects a view if it is the added element itself', async () => {
        const mutations: Observable<MutationRecordLike[]> = of([
            { addedNodes: [document.getElementById('1')!], removedNodes: [] },
        ])
        expect(
            await mutations
                .pipe(
                    trackViews([{ selector: '.view', resolveView: element => ({ element }) }]),
                    map(({ element }) => element.id),
                    toArray()
                )
                .toPromise()
        ).toEqual(['1'])
    })

    test('detects a view if it is the added element itself', async () => {
        const mutations: Observable<MutationRecordLike[]> = of([
            { addedNodes: [document.getElementById('1')!], removedNodes: [] },
        ])
        expect(
            await mutations
                .pipe(
                    trackViews([{ selector: '.view', resolveView: element => ({ element }) }]),
                    map(({ element }) => element.id),
                    toArray()
                )
                .toPromise()
        ).toEqual(['1'])
    })

    test('emits the element returned by the resolver', async () => {
        const mutations: Observable<MutationRecordLike[]> = of([{ addedNodes: [document.body], removedNodes: [] }])
        const selectorTarget = document.createElement('div')
        selectorTarget.className = 'selector-target'
        document.getElementById('1')!.append(selectorTarget)
        expect(
            await mutations
                .pipe(
                    trackViews([
                        {
                            selector: '.selector-target',
                            resolveView: element => ({ element: element.parentElement! }),
                        },
                    ]),
                    map(({ element }) => element.id),
                    toArray()
                )
                .toPromise()
        ).toEqual(['1'])
    })

    test("doesn't emit duplicate views", async () => {
        const mutations: Observable<MutationRecordLike[]> = of([{ addedNodes: [document.body], removedNodes: [] }])
        expect(
            await mutations
                .pipe(
                    trackViews([
                        { selector: '.view', resolveView: () => ({ element: document.getElementById('1')! }) },
                    ]),
                    map(({ element }) => element.id),
                    toArray()
                )
                .toPromise()
        ).toEqual(['1'])
    })

    test('detects views added later', async () => {
        const selector = '.test-code-view'
        const subscriber = sinon.spy()
        const mutations = new Subject<MutationRecordLike[]>()
        const { wait, done } = createBarrier()
        subscriptions.add(
            mutations
                .pipe(
                    trackViews([
                        {
                            selector,
                            resolveView: element => ({ element }),
                        },
                    ])
                )
                .subscribe(codeView => {
                    done()
                    subscriber(codeView)
                })
        )
        sinon.assert.notCalled(subscriber)
        mutations.next([{ addedNodes: [document.body], removedNodes: [] }])

        // Add code view to DOM
        const element = document.createElement('div')
        element.className = 'test-code-view'
        document.body.append(element)
        mutations.next([{ addedNodes: [element], removedNodes: [] }])
        await wait
        sinon.assert.calledOnce(subscriber)
        expect(subscriber.args[0].map(({ subscriptions, ...rest }) => rest)).toEqual([{ element }])
    })

    test('detects nested views added later', async () => {
        const selector = '.test-code-view'
        const subscriber = sinon.spy()
        const mutations = new Subject<MutationRecordLike[]>()
        const { wait, done } = createBarrier()
        subscriptions.add(
            mutations
                .pipe(
                    trackViews([
                        {
                            selector,
                            resolveView: element => ({ element }),
                        },
                    ])
                )
                .subscribe(codeView => {
                    done()
                    subscriber(codeView)
                })
        )
        sinon.assert.notCalled(subscriber)
        mutations.next([{ addedNodes: [document.body], removedNodes: [] }])

        // Add code view to DOM
        const element = document.createElement('div')
        element.className = 'test-code-view'
        const container = document.getElementById('parent')!
        container.append(element)
        mutations.next([{ addedNodes: [container], removedNodes: [] }])
        await wait
        sinon.assert.calledOnce(subscriber)
        expect(subscriber.args[0].map(({ subscriptions, ...rest }) => rest)).toEqual([{ element }])
    })

    test('removes views', async () => {
        const mutations = from<MutationRecordLike[][]>([
            [{ addedNodes: [document.body], removedNodes: [] }],
            [{ addedNodes: [], removedNodes: [document.getElementById('1')!] }],
            [{ addedNodes: [], removedNodes: [document.getElementById('3')!] }],
        ])
        await mutations
            .pipe(
                trackViews([{ selector: '.view', resolveView: element => ({ element }) }]),
                bufferCount(3),
                switchMap(async ([v1, v2, v3]) => {
                    const v2Removed = sinon.spy()
                    v2.subscriptions.add(v2Removed)
                    const v1Removed = new Promise(resolve => v1.subscriptions.add(resolve))
                    const v3Removed = new Promise(resolve => v3.subscriptions.add(resolve))
                    await Promise.all([v1Removed, v3Removed])
                    sinon.assert.notCalled(v2Removed)
                })
            )
            .toPromise()
    })

    test('removes all nested views', async () => {
        const mutations = from<MutationRecordLike[][]>([
            [{ addedNodes: [document.body], removedNodes: [] }],
            [{ addedNodes: [], removedNodes: [document.getElementById('parent')!] }],
        ])
        await mutations
            .pipe(
                trackViews([{ selector: '.view', resolveView: element => ({ element }) }]),
                bufferCount(3),
                switchMap(views =>
                    Promise.all(views.map(view => new Promise(resolve => view.subscriptions.add(resolve))))
                )
            )
            .toPromise()
    })

    test('removes a view without depending on its resolver', async () => {
        const selector = '.test-code-view'
        const subscriber = sinon.spy()
        const mutations = new Subject<MutationRecordLike[]>()
        const { wait, done } = createBarrier()

        // Track views using a resolver that looks at the element's parent tree
        // to determine whether it should resolve or return `null`.
        const resolver: ViewResolver<{ element: HTMLElement }> = {
            selector,
            resolveView: element => element.closest('.view') && { element },
        }
        subscriptions.add(
            mutations.pipe(trackViews([resolver])).subscribe(codeView => {
                done()
                subscriber(codeView)
            })
        )
        sinon.assert.notCalled(subscriber)
        mutations.next([{ addedNodes: [document.body], removedNodes: [] }])

        // Add code view to DOM
        const testElement = document.createElement('div')
        testElement.className = 'test-code-view'
        const container = document.getElementById('1')!
        container.append(testElement)
        mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
        await wait
        sinon.assert.calledOnce(subscriber)
        const view = subscriber.args[0][0] as { element: HTMLElement; subscriptions: Subscription }
        expect(view.element).toEqual(testElement)

        // Remove code view from the DOM. Verify it cannot be resolved anymore.
        testElement.remove()
        expect(resolver.resolveView(testElement)).toBe(null)

        // Verify that the code view still gets removed.
        const unsubscribed = new Promise(resolve => view.subscriptions.add(resolve))
        mutations.next([{ addedNodes: [], removedNodes: [testElement] }])
        await unsubscribed
    })
})
