import React from 'react'
import { GitCommitNodeByline } from './GitCommitNodeByline'
import { mount } from 'enzyme'

const FIXTURE_SIGNATURE_1 = {
    date: '1990-01-01',
    person: {
        name: 'Alice Zhao',
        displayName: 'Alice Zhao',
        email: 'alice@example.com',
        avatarURL: 'http://example.com/alice.png',
        user: null,
    },
}

const FIXTURE_SIGNATURE_2 = {
    date: '1991-01-01',
    person: {
        name: 'Bob Yang',
        displayName: 'Bob Yang',
        email: 'bob@example.com',
        avatarURL: 'http://example.com/bob.png',
        user: {
            username: 'bYang',
            displayName: 'Bob Yang',
            url: 'https://example.com/bobyang',
        },
    },
}

describe('GitCommitNodeByline', () => {
    test('author', () =>
        expect(mount(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />)).toMatchSnapshot())

    test('different author and committer', () =>
        expect(
            mount(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />)
        ).toMatchSnapshot())

    test('author (compact)', () =>
        expect(mount(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />)).toMatchSnapshot())

    test('different author and committer (compact)', () =>
        expect(
            mount(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />)
        ).toMatchSnapshot())

    test('same author and committer', () =>
        expect(
            mount(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_1} />)
        ).toMatchSnapshot())

    test('omit GitHub committer', () =>
        expect(
            mount(
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
                                displayName: 'Alice',
                                url: 'https://example.com',
                            },
                        },
                    }}
                />
            )
        ).toMatchSnapshot())
})
