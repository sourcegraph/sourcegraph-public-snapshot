import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import { ContributableMenu } from '../../api/protocol/contribution'
import { ActivationStatus } from './Activation'
import { ActivationPopoverButton } from './ActivationPopoverButton'

describe('ActivationPopoverButton', () => {
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => void 0 }
    const NOOP_EXTENSIONS_CONTROLLER = { services: {} as any, executeCommand: async () => void 0 }

    test('render 0/2', () => {
        const activation = new ActivationStatus(
            [
                {
                    id: 'id1',
                    title: 'title1',
                    detail: 'detail1',
                    action: (h: H.History) => void 0,
                },
                {
                    id: 'id2',
                    title: 'title2',
                    detail: 'detail2',
                    action: (h: H.History) => void 0,
                },
            ],
            () => of({})
        )
        const history = H.createMemoryHistory({ keyLength: 0 })
        const component = renderer.create(
            <ActivationPopoverButton
                history={history}
                activation={activation}
                location={history.location}
                menu={ContributableMenu.GlobalNav}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 1/2', () => {
        const activation = new ActivationStatus(
            [
                {
                    id: 'id1',
                    title: 'title1',
                    detail: 'detail1',
                    action: (h: H.History) => void 0,
                },
                {
                    id: 'id2',
                    title: 'title2',
                    detail: 'detail2',
                    action: (h: H.History) => void 0,
                },
            ],
            () => of({ id1: true })
        )
        const history = H.createMemoryHistory({ keyLength: 0 })
        const component = renderer.create(
            <ActivationPopoverButton
                history={history}
                activation={activation}
                location={history.location}
                menu={ContributableMenu.GlobalNav}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
