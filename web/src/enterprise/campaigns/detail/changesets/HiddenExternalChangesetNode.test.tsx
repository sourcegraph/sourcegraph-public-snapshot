import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'
import { addDays } from 'date-fns'
import { ChangesetState } from '../../../../../../shared/src/graphql/schema'

describe('HiddenExternalChangesetNode', () => {
    test('renders', () => {
        const renderer = createRenderer()
        renderer.render(
            <HiddenExternalChangesetNode
                node={{
                    id: 'test',
                    nextSyncAt: addDays(new Date(), 1).toISOString(),
                    state: ChangesetState.OPEN,
                    updatedAt: new Date().toISOString(),
                }}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
