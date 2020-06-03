import React from 'react'
import renderer from 'react-test-renderer'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import { subMinutes } from 'date-fns'

describe('ChangesetLastSynced', () => {
    test('renders not scheduled', () => {
        const result = renderer.create(
            <ChangesetLastSynced
                disableRefresh={false}
                changeset={{
                    id: '123',
                    nextSyncAt: null,
                    updatedAt: subMinutes(new Date('2020-03-01'), 10).toISOString(),
                }}
                _now={new Date('2020-03-01')}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })
    test('renders scheduled', () => {
        const result = renderer.create(
            <ChangesetLastSynced
                disableRefresh={false}
                changeset={{
                    id: '123',
                    nextSyncAt: new Date('2020-03-02').toISOString(),
                    updatedAt: subMinutes(new Date('2020-03-01'), 10).toISOString(),
                }}
                _now={new Date('2020-03-01')}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })

    test('renders with refresh disabled', () => {
        const result = renderer.create(
            <ChangesetLastSynced
                disableRefresh={true}
                changeset={{
                    id: '123',
                    nextSyncAt: new Date('2020-03-02').toISOString(),
                    updatedAt: subMinutes(new Date('2020-03-01'), 10).toISOString(),
                }}
                _now={new Date('2020-03-01')}
            />
        )
        expect(result.toJSON()).toMatchSnapshot()
    })
})
