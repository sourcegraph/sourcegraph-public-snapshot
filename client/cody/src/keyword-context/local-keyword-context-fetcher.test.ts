import { regexForTerms, userQueryToKeywordQuery } from './local-keyword-context-fetcher'

describe('keyword context', () => {
    it('query to regex', () => {
        const trials: {
            userQuery: string
            expRegex: string
        }[] = [
            {
                userQuery: 'Where is auth in Sourcegraph?',
                expRegex: '(?:where|auth|sourcegraph)',
            },
            {
                userQuery: 'saml auth handler',
                expRegex: '(?:saml|auth|handler)',
            },
            {
                userQuery: 'Where is the HTTP middleware defined in this codebase?',
                expRegex: '(?:where|http|middlewar|defin|codebas)',
            },
        ]
        for (const trial of trials) {
            const terms = userQueryToKeywordQuery(trial.userQuery)
            const regex = regexForTerms(...terms)
            expect(regex).toEqual(trial.expRegex)
        }
    })
})
