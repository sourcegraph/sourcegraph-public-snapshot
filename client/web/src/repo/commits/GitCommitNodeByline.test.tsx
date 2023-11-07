import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'

import type { SignatureFields } from '../../graphql-operations'

import { GitCommitNodeByline } from './GitCommitNodeByline'

const FIXTURE_SIGNATURE_1: SignatureFields = {
    date: '1990-01-01',
    person: {
        name: 'Alice Zhao',
        displayName: 'Alice Zhao',
        email: 'alice@example.com',
        avatarURL: 'http://example.com/alice.png',
        user: null,
    },
}

const FIXTURE_SIGNATURE_2: SignatureFields = {
    date: '1991-01-01',
    person: {
        name: 'Bob Yang',
        displayName: 'Bob Yang',
        email: 'bob@example.com',
        avatarURL: 'http://example.com/bob.png',
        user: {
            username: 'bYang',
            id: 'user123',
            displayName: 'Bob Yang',
            url: 'https://example.com/bobyang',
        },
    },
}

describe('GitCommitNodeByline', () => {
    test('author', () =>
        expect(
            render(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />).asFragment()
        ).toMatchSnapshot())

    test('different author and committer', () =>
        expect(
            render(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />).asFragment()
        ).toMatchSnapshot())

    test('author (compact)', () =>
        expect(
            render(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={null} />).asFragment()
        ).toMatchSnapshot())

    test('different author and committer (compact)', () =>
        expect(
            render(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_2} />).asFragment()
        ).toMatchSnapshot())

    test('same author and committer', () =>
        expect(
            render(<GitCommitNodeByline author={FIXTURE_SIGNATURE_1} committer={FIXTURE_SIGNATURE_1} />).asFragment()
        ).toMatchSnapshot())

    test('omit GitHub committer', () =>
        expect(
            render(
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
                                id: 'user123',
                                displayName: 'Alice',
                                url: 'https://example.com',
                            },
                        },
                    }}
                />
            ).asFragment()
        ).toMatchSnapshot())
})
