import { authStatusInit } from './protocol'
import { convertGitCloneURLToCodebaseName, validateAuthStatus } from './utils'

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
    const isEnterprise = true
    const verifiedEmail = true
    const codyEnabled = true
    // DOTCOM AND APP USERS
    test('returns initial auth status for invalid user', () => {
        const expected = authStatusInit
        expect(validateAuthStatus(!isEnterprise, '', !verifiedEmail, !codyEnabled, '')).toEqual(expected)
    })

    test('returns auth status for valid user with varified email on non-enterprise instance', () => {
        const expected = {
            ...authStatusInit,
            authenticated: true,
            hasVerifiedEmail: true,
            showInvalidAccessTokenError: false,
            requiresVerifiedEmail: true,
            siteHasCodyEnabled: true,
        }
        expect(validateAuthStatus(!isEnterprise, '1', verifiedEmail, codyEnabled, '')).toEqual(expected)
    })

    test('returns auth status for valid user without verified email on non-enterprise instance', () => {
        const expected = {
            ...authStatusInit,
            authenticated: true,
            hasVerifiedEmail: false,
            requiresVerifiedEmail: true,
            siteHasCodyEnabled: true,
        }
        expect(validateAuthStatus(!isEnterprise, '1', !verifiedEmail, !codyEnabled, '')).toEqual(expected)
    })

    // ENTERPRISE
    test('returns auth status for valid user on enterprise instance with Cody enabled', () => {
        const expected = {
            ...authStatusInit,
            authenticated: true,
            siteHasCodyEnabled: true,
        }
        expect(validateAuthStatus(isEnterprise, '1', verifiedEmail, codyEnabled, '')).toEqual(expected)
    })

    test('returns auth status for invalid user on enterprise instance with Cody enabled', () => {
        const expected = authStatusInit
        expect(validateAuthStatus(isEnterprise, '', verifiedEmail, codyEnabled, '')).toEqual(expected)
    })

    test('returns auth status for valid user on enterprise instance without Cody enabled', () => {
        const expected = {
            ...authStatusInit,
            authenticated: true,
            siteHasCodyEnabled: false,
        }
        expect(validateAuthStatus(isEnterprise, '1', !verifiedEmail, !codyEnabled, '')).toEqual(expected)
    })

    test('returns auth status for invalid user on enterprise instance without Cody enabled', () => {
        const expected = authStatusInit
        expect(validateAuthStatus(isEnterprise, '', verifiedEmail, !codyEnabled, '')).toEqual(expected)
    })
})
