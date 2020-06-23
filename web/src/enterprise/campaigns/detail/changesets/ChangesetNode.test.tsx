import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { ChangesetNode } from './ChangesetNode'
import { Subject } from 'rxjs'
import { Changeset } from '../../../../../../shared/src/graphql/schema'

describe('ChangesetNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders ExternalChangeset', () => {
        const renderer = createRenderer()
        renderer.render(
            <ChangesetNode
                isLightTheme={true}
                history={history}
                location={location}
                viewerCanAdminister={false}
                campaignUpdates={new Subject<void>()}
                node={{ __typename: 'ExternalChangeset' } as Changeset}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
    test('renders HiddenExternalChangeset', () => {
        const renderer = createRenderer()
        renderer.render(
            <ChangesetNode
                isLightTheme={true}
                history={history}
                location={location}
                viewerCanAdminister={false}
                campaignUpdates={new Subject<void>()}
                node={{ __typename: 'HiddenExternalChangeset' } as Changeset}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
