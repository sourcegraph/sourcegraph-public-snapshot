import React from 'react'
import renderer from 'react-test-renderer'
import { AddChangesetForm } from './AddChangesetForm'
import { createMemoryHistory } from 'history'

describe('AddChangesetForm', () => {
    test('renders the form', () => {
        expect(
            renderer
                .create(<AddChangesetForm campaignID="123" onAdd={() => undefined} history={createMemoryHistory()} />)
                .toJSON()
        ).toMatchSnapshot()
    })
})
