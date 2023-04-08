import { regexForTerms, userQueryToKeywordQuery } from './local-keyword-context-fetcher'

type FileInfo = {
    [filename: string]: {
        termCounts: { [stem: string]: number } // keyed off of stem
        sizeBytes: number
    }
}

function filesToFileMatches(files: FileInfo): {
    totalFiles: number
    fileTermCounts: { [filename: string]: { [stem: string]: number } }
    termTotalFiles: { [stem: string]: number }
} {
    const fileTermCounts: { [filename: string]: { [stem: string]: number } } = {}
    for (const [filename, { termCounts }] of Object.entries(files)) {
        fileTermCounts[filename] = termCounts
    }
    const termTotalFiles: { [stem: string]: number } = {}
    for (const [, { termCounts }] of Object.entries(files)) {
        for (const stem of Object.keys(termCounts)) {
            termTotalFiles[stem]++
        }
    }

    return {
        totalFiles: Object.keys(files).length,
        fileTermCounts,
        termTotalFiles,
    }
}

describe('keyword context', () => {
    it('query to regex', () => {
        const trials: {
            userQuery: string
            expRegex: string
        }[] = [
            {
                userQuery: `Where is auth in Sourcegraph?`,
                expRegex: `(?:where|auth|sourcegraph)`,
            },
            {
                userQuery: `saml auth handler`,
                expRegex: `(?:saml|auth|handler)`,
            },
            {
                userQuery: `Where is the HTTP middleware defined in this codebase?`,
                expRegex: `(?:where|http|middlewar|defin|codebas)`,
            },
        ]
        for (const trial of trials) {
            const terms = userQueryToKeywordQuery(trial.userQuery)
            const regex = regexForTerms(...terms)
            expect(regex).toEqual(trial.expRegex)
        }
    })
})
