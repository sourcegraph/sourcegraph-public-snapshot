import React from 'react'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'
import { addDays } from 'date-fns'
import { ChangesetState } from '../../../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'

describe('HiddenExternalChangesetNode', () => {
    test('renders', () => {
        expect(
            mount(
                <HiddenExternalChangesetNode
                    node={{
                        id: 'test',
                        nextSyncAt: addDays(new Date(), 1).toISOString(),
                        state: ChangesetState.OPEN,
                        updatedAt: new Date().toISOString(),
                    }}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
