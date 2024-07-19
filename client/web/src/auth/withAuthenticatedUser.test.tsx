import type { FunctionComponent } from 'react'

import { render, screen } from '@testing-library/react'
import { MemoryRouter, useLocation, useNavigate } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { AuthenticatedUserOnly, withAuthenticatedUser } from './withAuthenticatedUser'

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom')
    return {
        ...actual,
        useNavigate: vi.fn(),
        useLocation: vi.fn(),
    }
})

const MockComponent: FunctionComponent<{
    authenticatedUser: Pick<AuthenticatedUser, 'username'>
    foo: string
}> = ({ authenticatedUser, foo }) => (
    <div>
        Authenticated: {authenticatedUser.username} - {foo}
    </div>
)

const WrappedComponent: FunctionComponent<{
    authenticatedUser: Pick<AuthenticatedUser, 'username'> | null
    foo: string
}> = withAuthenticatedUser(MockComponent)

describe('withAuthenticatedUser', () => {
    it('renders the component when user is authenticated', () => {
        render(<WrappedComponent authenticatedUser={{ username: 'testuser' }} foo="bar" />, {
            wrapper: MemoryRouter,
        })
        expect(screen.getByText('Authenticated: testuser - bar')).toBeInTheDocument()
    })

    it('redirects to sign-in when user is not authenticated', () => {
        const mockNavigate = vi.fn()
        vi.mocked(useNavigate).mockReturnValue(mockNavigate)

        vi.mocked(useLocation).mockReturnValue({
            pathname: '/foo',
            hash: '#h',
            search: '?q',
            key: '',
            state: undefined,
        })

        const { container } = render(<WrappedComponent authenticatedUser={null} foo="bar" />, { wrapper: MemoryRouter })
        expect(container).toBeEmptyDOMElement()
        expect(mockNavigate).toHaveBeenCalledWith('/sign-in?returnTo=%2Ffoo%3Fq%23h', { replace: true })
    })
})

// Test of typechecking only. We want to make sure that it's possible to use withAuthenticatedUser
// with an arrow-expression component.
// @ts-ignore unused
const WrappedArrowComponent: FunctionComponent<{
    authenticatedUser: Pick<AuthenticatedUser, 'username'> | null
    foo: string
}> = withAuthenticatedUser<
    {
        authenticatedUser: Pick<AuthenticatedUser, 'username'>
        foo: string
    },
    Pick<AuthenticatedUser, 'username'>
>(({ authenticatedUser, foo }) => (
    <div>
        Authenticated: {authenticatedUser.username} - {foo}
    </div>
))
// @ts-ignore unused
const WrappedArrowComponentWithoutAuthenticatedUserProp: FunctionComponent<{
    authenticatedUser: Pick<AuthenticatedUser, 'username'> | null
    foo: string
}> = withAuthenticatedUser<{
    foo: string
    authenticatedUser: never
}>(({ foo }) => <div>{foo}</div>)

describe('AuthenticatedUserOnly', () => {
    it('renders the component when user is authenticated', () => {
        render(<AuthenticatedUserOnly authenticatedUser={{ id: 'u' }}>authed</AuthenticatedUserOnly>, {
            wrapper: MemoryRouter,
        })
        expect(screen.getByText('authed')).toBeInTheDocument()
    })

    it('redirects to sign-in when user is not authenticated', () => {
        const mockNavigate = vi.fn()
        vi.mocked(useNavigate).mockReturnValue(mockNavigate)

        vi.mocked(useLocation).mockReturnValue({
            pathname: '/foo',
            hash: '#h',
            search: '?q',
            key: '',
            state: undefined,
        })

        const { container } = render(<AuthenticatedUserOnly authenticatedUser={null}>authed</AuthenticatedUserOnly>, {
            wrapper: MemoryRouter,
        })
        expect(container).toBeEmptyDOMElement()
        expect(mockNavigate).toHaveBeenCalledWith('/sign-in?returnTo=%2Ffoo%3Fq%23h', { replace: true })
    })
})
