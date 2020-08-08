import React from 'react'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'
import { addDays } from 'date-fns'
import { mount } from 'enzyme'
import { ChangesetExternalState } from '../../../../graphql-operations'

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
