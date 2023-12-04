import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, test } from 'vitest'

import { PersonLink } from './PersonLink'

describe('PersonLink', () => {
    test('no display name, only email', () =>
        expect(
            render(
                <PersonLink
                    person={{ displayName: '', email: 'alice@example.com', user: null }}
                    className="a"
                    userClassName="b"
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('no user account', () =>
        expect(
            render(
                <PersonLink
                    person={{ displayName: 'alice', email: 'alice@example.com', user: null }}
                    className="a"
                    userClassName="b"
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('with user account', () =>
        expect(
            render(
                <MemoryRouter>
                    <PersonLink
                        person={{
                            displayName: 'Alice',
                            email: 'alice@example.com',
                            user: { username: 'alice', displayName: 'Alice Smith', url: 'u' },
                        }}
                        className="a"
                        userClassName="b"
                    />
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot())
})
