import { of, Subject } from 'rxjs'
import { toArray } from 'rxjs/operators'
import * as sinon from 'sinon'
import { MutationRecordLike } from '../../shared/util/dom'
import { CodeViewSpecWithOutSelector, FileInfo } from './code_intelligence'
import { trackCodeViews } from './code_views'

describe('code_views', () => {
    beforeEach(() => {
        document.body.innerHTML = ''
    })
    describe('trackCodeViews()', () => {
        const fileInfo: FileInfo = {
            repoName: 'foo',
            filePath: '/bar.ts',
            commitID: '1',
        }
        const codeViewSpec: CodeViewSpecWithOutSelector = {
            dom: {
                getCodeElementFromTarget: () => null,
                getCodeElementFromLineNumber: () => null,
                getLineNumberFromCodeElement: () => 1,
            },
            resolveFileInfo: () => of(fileInfo),
        }
        it('should detect added code views from specs', async () => {
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            const selector = '.test-code-view'
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewSpecs: [{ selector, ...codeViewSpec }],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected).toEqual([{ ...codeViewSpec, codeViewElement, type: 'added' }])
        })
        it('should detect added code views from resolver', async () => {
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            const selector = '.test-code-view'
            const codeViewSpecResolver = { selector, resolveViewSpec: sinon.spy(() => codeViewSpec) }
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({ codeViewSpecResolver }),
                    toArray()
                )
                .toPromise()
            expect(detected).toEqual([{ ...codeViewSpec, codeViewElement, type: 'added' }])
            sinon.assert.calledOnce(codeViewSpecResolver.resolveViewSpec)
            sinon.assert.calledWith(codeViewSpecResolver.resolveViewSpec, codeViewElement)
        })
        it('should detect an added code view if it is the added element itself', async () => {
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            const selector = '.test-code-view'
            const detected = await of([{ addedNodes: [codeViewElement], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewSpecs: [{ selector, ...codeViewSpec }],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected).toEqual([{ ...codeViewSpec, codeViewElement, type: 'added' }])
        })
        it('should detect added code views added later', async () => {
            const selector = '.test-code-view'
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            mutations.pipe(trackCodeViews({ codeViewSpecs: [{ selector, ...codeViewSpec }] })).subscribe(subscriber)
            sinon.assert.notCalled(subscriber)
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])

            // Add code view to DOM
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            mutations.next([{ addedNodes: [codeViewElement], removedNodes: [] }])
            sinon.assert.calledOnce(subscriber)
            expect(subscriber.args[0]).toEqual([{ ...codeViewSpec, codeViewElement, type: 'added' }])
        })
        it('should detect nested added code views added later', async () => {
            const selector = '.test-code-view'
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            mutations.pipe(trackCodeViews({ codeViewSpecs: [{ selector, ...codeViewSpec }] })).subscribe(subscriber)
            sinon.assert.notCalled(subscriber)
            mutations.next([{ addedNodes: [], removedNodes: [] }])

            // Add code view to DOM
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
            sinon.assert.calledOnce(subscriber)
            expect(subscriber.args[0]).toEqual([{ ...codeViewSpec, codeViewElement, type: 'added' }])
        })
        it('should detect removed code views', async () => {
            const selector = '.test-code-view'
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            document.body.append(codeViewElement)
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            mutations.pipe(trackCodeViews({ codeViewSpecs: [{ selector, ...codeViewSpec }] })).subscribe(subscriber)
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
            sinon.assert.calledOnce(subscriber)

            // Remove code view from DOM
            codeViewElement.remove()
            mutations.next([{ addedNodes: [], removedNodes: [codeViewElement] }])
            sinon.assert.calledTwice(subscriber)
            expect(subscriber.args).toEqual([
                [{ ...codeViewSpec, codeViewElement, type: 'added' }],
                [{ codeViewElement, type: 'removed' }],
            ])
        })
        it('should detect nested removed code views', async () => {
            const selector = '.test-code-view'
            const codeViewElement = document.createElement('div')
            codeViewElement.className = 'test-code-view'
            const container = document.body.appendChild(document.createElement('div'))
            container.append(codeViewElement)
            const subscriber = sinon.spy()
            const mutations = new Subject<MutationRecordLike[]>()
            mutations.pipe(trackCodeViews({ codeViewSpecs: [{ selector, ...codeViewSpec }] })).subscribe(subscriber)
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
            sinon.assert.calledOnce(subscriber)

            // Remove code view from DOM
            container.remove()
            mutations.next([{ addedNodes: [], removedNodes: [container] }])
            sinon.assert.calledTwice(subscriber)
            expect(subscriber.args).toEqual([
                [{ ...codeViewSpec, codeViewElement, type: 'added' }],
                [{ codeViewElement, type: 'removed' }],
            ])
        })
    })
})
