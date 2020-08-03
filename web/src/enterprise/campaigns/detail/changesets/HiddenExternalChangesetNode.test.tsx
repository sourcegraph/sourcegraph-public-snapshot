import React from 'react'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'
import { addDays } from 'date-fns'
import { ChangesetExternalState } from '../../../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'

describe('HiddenExternalChangesetNode', () => {
    test('renders', () => {
        expect(
            mount(
                <HiddenExternalChangesetNode
                    node={{
                        id: 'test',
                        nextSyncAt: addDays(new Date(), 1).toISOString(),
                        externalState: ChangesetExternalState.OPEN,
                        updatedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot()
    })
})
