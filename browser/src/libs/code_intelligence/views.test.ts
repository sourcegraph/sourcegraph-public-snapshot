import { from, Observable, of } from 'rxjs'
import { bufferCount, switchMap, toArray } from 'rxjs/operators'
import * as sinon from 'sinon'
import { MutationRecordLike } from '../../shared/util/dom'
import { trackViews } from './views'

const FIXTURE_HTML = `
    <div id="parent">
        <div class="view" id="1"></div>
        <div class="view" id="2"></div>
        <div class="view" id="3"></div>
    </div>
`

describe('trackViews()', () => {
    beforeEach(() => {
        document.body.innerHTML = FIXTURE_HTML
    })

    afterAll(() => {
        document.body.innerHTML = ''
    })

    test('emits all views on the page', async () => {
        const mutations: Observable<MutationRecordLike[]> = of([{ addedNodes: [document.body], removedNodes: [] }])
        const views = await mutations
            .pipe(
                trackViews([{ selector: '.view', resolveView: element => ({ element }) }]),
                toArray()
            )
            .toPromise()
        expect(views.map(({ element }) => element.id)).toEqual(['1', '2', '3'])
    })

    test("doesn't emit duplicate views", async () => {
        const mutations: Observable<MutationRecordLike[]> = of([{ addedNodes: [document.body], removedNodes: [] }])
        const views = await mutations
            .pipe(
                trackViews([{ selector: '.view', resolveView: () => ({ element: document.getElementById('1')! }) }]),
                toArray()
            )
            .toPromise()
        expect(views.map(({ element }) => element.id)).toEqual(['1'])
    })

    test('removes a view when its element gets removed', async () => {
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

    test('removes all views contained within a removed element', async () => {
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
})
