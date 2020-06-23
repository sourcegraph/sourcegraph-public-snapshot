import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { PatchInterfaceNode } from './PatchInterfaceNode'
import { Subject } from 'rxjs'
import { PatchInterface } from '../../../../../../shared/src/graphql/schema'

describe('PatchInterfaceNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders HiddenPatch', () => {
        const renderer = createRenderer()
        renderer.render(
            <PatchInterfaceNode
                isLightTheme={true}
                history={history}
                location={location}
                campaignUpdates={new Subject<void>()}
                enablePublishing={false}
                node={{ __typename: 'HiddenPatch' } as PatchInterface}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
    test('renders Patch', () => {
        const renderer = createRenderer()
        renderer.render(
            <PatchInterfaceNode
                isLightTheme={true}
                history={history}
                location={location}
                campaignUpdates={new Subject<void>()}
                enablePublishing={false}
                node={{ __typename: 'Patch' } as PatchInterface}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
