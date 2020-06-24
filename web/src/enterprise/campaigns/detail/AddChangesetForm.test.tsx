import React from 'react'
import { AddChangesetForm } from './AddChangesetForm'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

describe('AddChangesetForm', () => {
    test('renders the form', () => {
        expect(
            mount(<AddChangesetForm campaignID="123" onAdd={() => undefined} history={createMemoryHistory()} />)
        ).toMatchSnapshot()
    })
})
