import * as H from 'history'
import React from 'react'
import { PatchInterfaceNode } from './PatchInterfaceNode'
import { Subject } from 'rxjs'
import { PatchInterface } from '../../../../../../shared/src/graphql/schema'
import { shallow } from 'enzyme'

describe('PatchInterfaceNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders HiddenPatch', () => {
        expect(
            shallow(
                <PatchInterfaceNode
                    isLightTheme={true}
                    history={history}
                    location={location}
                    campaignUpdates={new Subject<void>()}
                    enablePublishing={false}
                    node={{ __typename: 'HiddenPatch' } as PatchInterface}
                />
            )
        ).toMatchSnapshot()
    })
    test('renders Patch', () => {
        expect(
            shallow(
                <PatchInterfaceNode
                    isLightTheme={true}
                    history={history}
                    location={location}
                    campaignUpdates={new Subject<void>()}
                    enablePublishing={false}
                    node={{ __typename: 'Patch' } as PatchInterface}
                />
            )
        ).toMatchSnapshot()
    })
})
