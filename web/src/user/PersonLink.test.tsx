import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'
import { PersonLink } from './PersonLink'

describe('PersonLink', () => {
    test('no user account', () =>
        expect(renderer.create(<PersonLink user="alice" className="a" userClassName="b" />).toJSON()).toMatchSnapshot())

    test('with user account', () =>
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <PersonLink
                            user={{ displayName: 'Alice', username: 'alice', url: 'u' }}
                            className="a"
                            userClassName="b"
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot())
})
