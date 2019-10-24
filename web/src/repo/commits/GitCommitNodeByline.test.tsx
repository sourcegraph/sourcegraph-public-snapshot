import React from 'react'
import renderer from 'react-test-renderer'
import { GitCommitNodeByline, Signature } from './GitCommitNodeByline'

const FIXTURE_SIGNATURE_1: Signature = {
    date: '1990-01-01',
    person: {
        name: 'Alice Zhao',
        displayName: 'Alice Zhao',
        email: 'alice@example.com',
        avatarURL: 'http://example.com/alice.png',
        user: null,
    },
}

const FIXTURE_SIGNATURE_2: Signature = {
    date: '1991-01-01',
    person: {
        name: 'Bob Yang',
        displayName: 'Bob Yang',
        email: 'bob@example.com',
        avatarURL: 'http://example.com/bob.png',
        user: {
            username: 'bYang',
        },
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

    test('omit GitHub committer', () =>
        expect(
            renderer
                .create(
                    <GitCommitNodeByline
                        author={FIXTURE_SIGNATURE_1}
                        committer={{
                            date: '1992-01-01',
                            person: {
                                name: 'GitHub',
                                email: 'noreply@github.com',
                                displayName: 'GitHub',
                                avatarURL: 'http://example.com/github.png',
                                user: {
                                    username: 'gitUserName',
                                },
                            },
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
