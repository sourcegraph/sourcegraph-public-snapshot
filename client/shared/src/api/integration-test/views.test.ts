import { NEVER, of } from 'rxjs'
import { first, take, toArray } from 'rxjs/operators'
import { wrapRemoteObservable } from '../client/api/common'
import { ContributableViewContainer } from '../protocol'
import { assertToJSON, integrationTestContext } from './testHelpers'

describe('Views (integration)', () => {
    describe('app.createPanelView', () => {
        test('no component', async () => {
            const { extensionAPI, extensionHostAPI } = await integrationTestContext()
            const panelView = extensionAPI.app.createPanelView('p')
            panelView.title = 't'
            panelView.content = 'c'
            panelView.priority = 3

            const values = await wrapRemoteObservable(extensionHostAPI.getPanelViews())
                .pipe(first(views => views.length > 0))
                .toPromise()

            assertToJSON(values, [
                {
                    id: 'p',
                    title: 't',
                    content: 'c',
                    priority: 3,
                    component: null,
                },
            ])
        })

        test('with component (location provider)', async () => {
            const { extensionAPI, extensionHostAPI } = await integrationTestContext()

            const LOCATION_PROVIDER_ID = 'x'
            extensionAPI.languages.registerLocationProvider(LOCATION_PROVIDER_ID, ['*'], {
                provideLocations: () => NEVER,
            })

            const panelView = extensionAPI.app.createPanelView('p')
            panelView.title = 't'
            panelView.content = 'c'
            panelView.priority = 3
            panelView.component = { locationProvider: LOCATION_PROVIDER_ID }

            const values = await wrapRemoteObservable(extensionHostAPI.getPanelViews())
                .pipe(first(views => views.length > 0))
                .toPromise()

            assertToJSON(values, [
                {
                    id: 'p',
                    title: 't',
                    content: 'c',
                    priority: 3,
                    component: {
                        locationProvider: LOCATION_PROVIDER_ID,
                    },
                },
            ])
        })
    })

    test('app.registerViewProvider', async () => {
        const { extensionAPI, extensionHostAPI } = await integrationTestContext()

        extensionAPI.app.registerViewProvider('v', {
            where: ContributableViewContainer.GlobalPage,
            provideView: parameters => of({ title: `t${parameters.x}`, content: [] }),
        })

        const views = await wrapRemoteObservable(extensionHostAPI.getGlobalPageViews({ x: 'y' }))
            .pipe(take(2), toArray())
            .toPromise()
        expect(views).toEqual([[{ id: 'v', view: undefined }], [{ id: 'v', view: { title: 'ty', content: [] } }]])
    })
})
