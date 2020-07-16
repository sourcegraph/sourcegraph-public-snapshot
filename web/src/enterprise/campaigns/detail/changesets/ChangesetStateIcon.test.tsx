import React from 'react'
import { ChangesetStateIcon } from './ChangesetStateIcon'
import { ChangesetExternalState } from '../../../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'

describe('ChangesetStateIcon', () => {
    for (const state of Object.values(ChangesetExternalState)) {
        test(`renders ${state}`, () => {
            expect(mount(<ChangesetStateIcon externalState={state} />)).toMatchSnapshot()
        })
    }
})
