import { convertGitCloneURLToCodebaseName } from './ChatViewProvider'

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
