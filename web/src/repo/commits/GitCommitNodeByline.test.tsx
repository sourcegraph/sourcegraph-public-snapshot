import React from 'react'
import renderer from 'react-test-renderer'
import * as GQL from '../../../../shared/src/graphql/schema'
import { GitCommitNodeByline } from './GitCommitNodeByline'

const FIXTURE_SIGNATURE_1: GQL.ISignature = {
    __typename: 'Signature',
    date: '1990-01-01',
    person: {
        __typename: 'Person',
        name: 'Alice Zhao',
        displayName: 'Alice Zhao',
        email: 'alice@example.com',
        avatarURL: 'http://example.com/alice.png',
        user: null,
    },
}

const FIXTURE_SIGNATURE_2: GQL.ISignature = {
    __typename: 'Signature',
    date: '1991-01-01',
    person: {
        __typename: 'Person',
        name: 'Bob Yang',
        displayName: 'Bob Yang',
        email: 'bob@example.com',
        avatarURL: 'http://example.com/bob.png',
        user: null,
    },
}

describe('GitCommitNodeByline', () => {
    test('author', () =>
        expect(
            renderer.create(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />).toJSON()
        ).toMatchSnapshot())

    test('different author and committer', () =>
        expect(
            renderer
                .create(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />)
                .toJSON()
        ).toMatchSnapshot())

    test('author (compact)', () =>
        expect(
            renderer.create(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />).toJSON()
        ).toMatchSnapshot())

    test('different author and committer (compact)', () =>
        expect(
            renderer
                .create(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />)
                .toJSON()
        ).toMatchSnapshot())

    test('same author and committer', () =>
        expect(
            renderer
                .create(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_1} />)
                .toJSON()
        ).toMatchSnapshot())
})
