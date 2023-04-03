import { find as linkifyFind } from 'linkifyjs'

import { ExternalServiceKind } from '../../graphql-operations'

// Some places return an URL to objects within a repo, e.g.:
//
// https://github.com/sourcegraph/sourcegraph/commit/ad1ea519e5a31bb868be947107bcf43f4f9fc672
//
// This function removes those unwanted parts
const GITHUB_URL_SCHEMA = /^(https?:\/\/[^/]+\/[^/]+\/[^/]+)(.*)$/
function githubRepoUrl(url: string): string {
    const match = url.match(GITHUB_URL_SCHEMA)
    if (match?.[1]) {
        return match[1]
    }

    return url
}

// This regex is supposed to match in the following cases:
//
//  - Create search and search-ui packages (#29773)
//  - Fix #123 for xyz
//
// However it is supposed not to match in:
//
// - Something sourcegraph/other-repo#123 or so
// - 123#123
const GH_ISSUE_NUMBER_IN_COMMIT = /([^\dA-Za-z](#\d+))/g

const getGitHubIssueLinks = (input: string, externalServiceUrl: string): LinkFromString[] => {
    const links = []

    const matches = [...input.matchAll(GH_ISSUE_NUMBER_IN_COMMIT)]
    if (matches.length > 0) {
        const url = githubRepoUrl(externalServiceUrl)
        for (const match of matches) {
            if (match.index === undefined) {
                continue
            }
            const issueNumber = match[2]
            links.push({
                start: match.index + 1,
                end: match.index + match[0].length,
                href: `${url}/pull/${issueNumber.replace('#', '')}`,
                value: issueNumber,
                type: 'gh-issue' as const,
            })
        }
    }

    return links
}

/**
 * Note: Matching URLs within a random string is difficult, as a URL can contain almost any character.
 * For example, it is valid to end a URL with parentheses or other punctuation, but in most cases this will not be desired.
 * We use linkifyjs to capture these edge cases and focus on the most common URLs.
] */
const getLinks = (input: string): LinkFromString[] => {
    const links = linkifyFind(input)
    return links
        .filter(({ value }) =>
            // Filter out links that don't begin with a protocol.
            // This ensures we don't accidentally parse file names as links.
            /^(https?|ftp|file):\/\//.test(value)
        )
        .map(link => ({
            start: link.start,
            end: link.end,
            href: link.href,
            value: link.value,
            type: 'url',
        }))
}

interface GetLinksFromStringParams {
    input: string
    externalURLs?: { url: string; serviceKind: ExternalServiceKind | null }[]
}

interface LinkFromString {
    start: number
    end: number
    href: string
    value: string
    type: 'url' | 'gh-issue'
}

/**
 * Given an input string, returns a sorted array of links found within the string.
 * If `externalURLs` is provided, GitHub issue references (e.g. #1234) will be parsed and included as links.
 */
export const getLinksFromString = ({ input, externalURLs }: GetLinksFromStringParams): LinkFromString[] => {
    const github = externalURLs ? externalURLs.find(url => url.serviceKind === ExternalServiceKind.GITHUB) : null
    const githubLinks = github ? getGitHubIssueLinks(input, github.url) : []

    const links = [...getLinks(input), ...githubLinks]
        .sort((a, b) => a.start - b.start)
        .filter((link, index, links) => {
            // Filter out links that are contained within another link.
            // This avoids a scenario where a link is rendered twice, once as a URL and once as a GH issue.
            if (index === 0) {
                return true
            }
            return link.start >= links[index - 1].end
        })

    return links
}
