import { defaultAuthStatus, unauthenticatedStatus } from './protocol'
import { convertGitCloneURLToCodebaseName, newAuthStatus } from './utils'

describe('convertGitCloneURLToCodebaseName', () => {
    test('converts GitHub SSH URL', () => {
        expect(convertGitCloneURLToCodebaseName('git@github.com:sourcegraph/sourcegraph.git')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('converts GitHub SSH URL no trailing .git', () => {
        expect(convertGitCloneURLToCodebaseName('git@github.com:sourcegraph/sourcegraph')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('converts GitHub HTTPS URL', () => {
        expect(convertGitCloneURLToCodebaseName('https://github.com/sourcegraph/sourcegraph')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('converts Bitbucket HTTPS URL', () => {
        expect(convertGitCloneURLToCodebaseName('https://username@bitbucket.org/sourcegraph/sourcegraph.git')).toEqual(
            'bitbucket.org/sourcegraph/sourcegraph'
        )
    })

    test('converts Bitbucket SSH URL', () => {
        expect(convertGitCloneURLToCodebaseName('git@bitbucket.sgdev.org:sourcegraph/sourcegraph.git')).toEqual(
            'bitbucket.sgdev.org/sourcegraph/sourcegraph'
        )
    })

    test('converts GitLab SSH URL', () => {
        expect(convertGitCloneURLToCodebaseName('git@gitlab.com:sourcegraph/sourcegraph.git')).toEqual(
            'gitlab.com/sourcegraph/sourcegraph'
        )
    })

    test('converts GitLab HTTPS URL', () => {
        expect(convertGitCloneURLToCodebaseName('https://gitlab.com/sourcegraph/sourcegraph.git')).toEqual(
            'gitlab.com/sourcegraph/sourcegraph'
        )
    })

    test('converts GitHub SSH URL with Git', () => {
        expect(convertGitCloneURLToCodebaseName('git@github.com:sourcegraph/sourcegraph.git')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('converts Eriks SSH Alias URL', () => {
        expect(convertGitCloneURLToCodebaseName('github:sourcegraph/sourcegraph')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('converts HTTP URL', () => {
        expect(convertGitCloneURLToCodebaseName('http://github.com/sourcegraph/sourcegraph')).toEqual(
            'github.com/sourcegraph/sourcegraph'
        )
    })

    test('returns null for invalid URL', () => {
        expect(convertGitCloneURLToCodebaseName('invalid')).toEqual(null)
    })
})

describe('validateAuthStatus', () => {
    // NOTE: Site version is for frontend use and doesn't play a role in validating auth status
    const siteVersion = ''
    const isDotComOrApp = true
    const verifiedEmail = true
    const codyEnabled = true
    const validUser = true
    const endpoint = 'https://example.com'
    // DOTCOM AND APP USERS
    test('returns auth state for invalid user on dotcom or app instance', () => {
        const expected = { ...unauthenticatedStatus, endpoint }
        expect(newAuthStatus(endpoint, isDotComOrApp, !validUser, !verifiedEmail, codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    test('returns auth status for valid user with varified email on dotcom or app instance', () => {
        const expected = {
            ...defaultAuthStatus,
            authenticated: true,
            hasVerifiedEmail: true,
            showInvalidAccessTokenError: false,
            requiresVerifiedEmail: true,
            siteHasCodyEnabled: true,
            isLoggedIn: true,
            endpoint,
        }
        expect(newAuthStatus(endpoint, isDotComOrApp, validUser, verifiedEmail, codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    test('returns auth status for valid user without verified email on dotcom or app instance', () => {
        const expected = {
            ...defaultAuthStatus,
            authenticated: true,
            hasVerifiedEmail: false,
            requiresVerifiedEmail: true,
            siteHasCodyEnabled: true,
            endpoint,
        }
        expect(newAuthStatus(endpoint, isDotComOrApp, validUser, !verifiedEmail, codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    // ENTERPRISE
    test('returns auth status for valid user on enterprise instance with Cody enabled', () => {
        const expected = {
            ...defaultAuthStatus,
            authenticated: true,
            siteHasCodyEnabled: true,
            isLoggedIn: true,
            endpoint,
        }
        expect(newAuthStatus(endpoint, !isDotComOrApp, validUser, verifiedEmail, codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    test('returns auth status for invalid user on enterprise instance with Cody enabled', () => {
        const expected = { ...unauthenticatedStatus, endpoint }
        expect(newAuthStatus(endpoint, !isDotComOrApp, !validUser, verifiedEmail, codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    test('returns auth status for valid user on enterprise instance with Cody disabled', () => {
        const expected = {
            ...defaultAuthStatus,
            authenticated: true,
            siteHasCodyEnabled: false,
            endpoint,
        }
        expect(newAuthStatus(endpoint, !isDotComOrApp, validUser, !verifiedEmail, !codyEnabled, siteVersion)).toEqual(
            expected
        )
    })

    test('returns auth status for invalid user on enterprise instance with Cody disabled', () => {
        const expected = { ...unauthenticatedStatus, endpoint }
        expect(newAuthStatus(endpoint, !isDotComOrApp, !validUser, verifiedEmail, !codyEnabled, siteVersion)).toEqual(
            expected
        )
    })
})
