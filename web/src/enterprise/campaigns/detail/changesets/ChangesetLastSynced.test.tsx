import React from 'react'
import renderer from 'react-test-renderer'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import { subMinutes } from 'date-fns'

describe('ChangesetLastSynced', () => {
    test('renders', () => {
        const result = renderer.create(
            <ChangesetLastSynced
                changeset={{
                    id: '123',
                    updatedAt: subMinutes(new Date('2020-03-01'), 10).toISOString(),
                }}
                _now={new Date('2020-03-01')}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })
})
