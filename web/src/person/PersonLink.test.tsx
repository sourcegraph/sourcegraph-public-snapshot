import React from 'react'
import { MemoryRouter } from 'react-router'
import { PersonLink } from './PersonLink'
import { mount } from 'enzyme'

describe('PersonLink', () => {
    test('no user account', () =>
        expect(
            mount(
                <PersonLink
                    person={{ displayName: 'alice', email: 'alice@example.com', user: null }}
                    className="a"
                    userClassName="b"
                />
            )
        ).toMatchSnapshot())

    test('with user account', () =>
        expect(
            mount(
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
            )
        ).toMatchSnapshot())
})
