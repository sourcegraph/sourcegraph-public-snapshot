import React from 'react'
import { ChangesetStateIcon } from './ChangesetStateIcon'
import { mount } from 'enzyme'
import { ChangesetExternalState } from '../../../../graphql-operations'

describe('ChangesetStateIcon', () => {
    for (const state of Object.values(ChangesetExternalState)) {
        test(`renders ${state}`, () => {
            expect(mount(<ChangesetStateIcon externalState={state} />)).toMatchSnapshot()
        })
    }
})
