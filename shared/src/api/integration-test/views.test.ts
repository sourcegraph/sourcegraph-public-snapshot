import { NEVER } from 'rxjs'
import { first } from 'rxjs/operators'
import { ContributableViewContainer } from '../protocol'
import { assertToJSON, collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Views (integration)', () => {
    describe('app.createPanelView', () => {
        test('no component', async () => {
            const { extensionAPI, services } = await integrationTestContext()
            const panelView = extensionAPI.app.createPanelView('p')
            panelView.title = 't'
            panelView.content = 'c'
            panelView.priority = 3
            await extensionAPI.internal.sync()

            const values = collectSubscribableValues(
                services.views.getViews(ContributableViewContainer.Panel).pipe(first())
            )
            assertToJSON(values, [
                [
                    {
                        id: 'p',
                        title: 't',
                        content: 'c',
                        priority: 3,
                        container: ContributableViewContainer.Panel,
                    },
                ],
            ])
        })

        test('with component (location provider)', async () => {
            const { extensionAPI, services } = await integrationTestContext()

            const LOCATION_PROVIDER_ID = 'x'
            extensionAPI.languages.registerLocationProvider(LOCATION_PROVIDER_ID, ['*'], {
                provideLocations: () => NEVER,
            })

            const panelView = extensionAPI.app.createPanelView('p')
            panelView.title = 't'
            panelView.content = 'c'
            panelView.priority = 3
            panelView.component = { locationProvider: LOCATION_PROVIDER_ID }
            await extensionAPI.internal.sync()

            const values = collectSubscribableValues(
                services.views.getViews(ContributableViewContainer.Panel).pipe(first())
            )
            assertToJSON(values.map(v => v.map(v => ({ ...v, locationProvider: 'value not checked' }))), [
                [
                    {
                        id: 'p',
                        title: 't',
                        content: 'c',
                        priority: 3,
                        container: ContributableViewContainer.Panel,
                        locationProvider: 'value not checked',
                    },
                ],
            ])
        })
    })
})
