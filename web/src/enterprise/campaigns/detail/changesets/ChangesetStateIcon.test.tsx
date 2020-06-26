import React from 'react'
import { ChangesetStateIcon } from './ChangesetStateIcon'
import { ChangesetState } from '../../../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'

describe('ChangesetStateIcon', () => {
    for (const state of Object.values(ChangesetState)) {
        test(`renders ${state}`, () => {
            expect(mount(<ChangesetStateIcon state={state} />).children()).toMatchSnapshot()
        })
    }
})
