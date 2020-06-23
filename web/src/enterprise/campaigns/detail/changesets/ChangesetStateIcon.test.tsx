import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { ChangesetStateIcon } from './ChangesetStateIcon'
import { ChangesetState } from '../../../../../../shared/src/graphql/schema'

describe('ChangesetStateIcon', () => {
    for (const state of Object.values(ChangesetState)) {
        test(`renders ${state}`, () => {
            const renderer = createRenderer()
            renderer.render(<ChangesetStateIcon state={state} />)
            expect(renderer.getRenderOutput().props).toMatchSnapshot()
        })
    }
})
