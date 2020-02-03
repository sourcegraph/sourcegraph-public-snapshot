import React from 'react'
import renderer from 'react-test-renderer'
import { ChangesetLabel } from './ChangesetLabel'

describe('ChangesetLabel', () => {
    test('renders a light label with dark text', () => {
        const result = renderer.create(
            <ChangesetLabel
                label={{
                    __typename: 'ChangesetLabel',
                    text: 'bug',
                    description: 'Something is wrong',
                    color: 'acfc99',
                }}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })
    test('renders a dark label with white text', () => {
        const result = renderer.create(
            <ChangesetLabel
                label={{
                    __typename: 'ChangesetLabel',
                    text: 'bug',
                    description: 'Something is wrong',
                    color: '330912',
                }}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })
})
