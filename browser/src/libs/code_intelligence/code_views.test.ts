import { from, of, Subject, Subscription } from 'rxjs'
import { switchMap, toArray } from 'rxjs/operators'
import * as sinon from 'sinon'
import { Omit } from 'utility-types'
import { createBarrier } from '../../../../shared/src/api/integration-test/testHelpers'
import { MutationRecordLike } from '../../shared/util/dom'
import { FileInfo } from './code_intelligence'
import { CodeView, toCodeViewResolver, trackCodeViews } from './code_views'

describe('code_views', () => {
    let subscriptions = new Subscription()
    beforeEach(() => {
        document.body.innerHTML = ''
    })
    afterEach(() => {
        subscriptions.unsubscribe()
        subscriptions = new Subscription()
    })
    describe('trackCodeViews()', () => {
        const fileInfo: FileInfo = {
            repoName: 'foo',
            filePath: '/bar.ts',
            commitID: '1',
        }
        const codeViewSpec: Omit<CodeView, 'element'> = {
            dom: {
                getCodeElementFromTarget: () => null,
                getCodeElementFromLineNumber: () => null,
                getLineElementFromLineNumber: () => null,
                getLineNumberFromCodeElement: () => 1,
            },
            resolveFileInfo: () => of(fileInfo),
        }
        it('should detect added code views from specs', async () => {
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            const selector = '.test-code-view'
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected.map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
        })
        it('should detect added code views from resolver', async () => {
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            const selector = '.test-code-view'
            const resolveView = sinon.spy((element: HTMLElement) => ({ element, ...codeViewSpec }))
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [{ selector, resolveView }],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected.map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
            sinon.assert.calledOnce(resolveView)
            sinon.assert.calledWith(resolveView, element)
        })
        it('should detect an added code view if it is the added element itself', async () => {
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            const selector = '.test-code-view'
            const detected = await of([{ addedNodes: [element], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected.map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
        })

        it('should detect added code views added later', async () => {
            const selector = '.test-code-view'
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            const { wait, done } = createBarrier()
            subscriptions.add(
                mutations
                    .pipe(
                        trackCodeViews({
                            codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                        })
                    )
                    .subscribe(codeViews => {
                        subscriber(codeViews)
                        done()
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
            expect(subscriber.args[0].map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
        })

        it('should detect nested added code views added later', async () => {
            const selector = '.test-code-view'
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            const { wait, done } = createBarrier()
            mutations
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                    })
                )
                .subscribe(codeView => {
                    done()
                    subscriber(codeView)
                })
            sinon.assert.notCalled(subscriber)
            mutations.next([{ addedNodes: [], removedNodes: [] }])

            // Add code view to DOM
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
            await wait
            sinon.assert.calledOnce(subscriber)
            expect(subscriber.args[0].map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
        })
        it('should detect nested removed code views', async () => {
            const selector = '.test-code-view'
            const element = document.createElement('div')
            element.className = 'test-code-view'
            const container = document.body.appendChild(document.createElement('div'))
            container.append(element)
            const mutations = from<MutationRecordLike[][]>([
                [{ addedNodes: [document.body], removedNodes: [] }],
                [{ addedNodes: [], removedNodes: [container] }],
            ])
            await mutations
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                    }),
                    switchMap(async view => new Promise(resolve => view.subscriptions.add(resolve)))
                )
                .toPromise()
        })
    })
})
