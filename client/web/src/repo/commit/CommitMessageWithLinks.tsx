import { Link } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../graphql-operations'

// This regex is supposed to match in the following cases:
//
//  - Create search and search-ui packages (#29773)
//  - Fix #123 for xyz
//
// However it is supposed not to mach in:
//
// - Something sourcegraph/other-repo#123 or so
// - 123#123
const GH_ISSUE_NUMBER_IN_COMMIT = /([^\dA-Za-z](#\d+))/g

interface Props {
    message: string
    to: string
    className: string
    onClick?: () => void
    externalURLs: { url: string; serviceKind: ExternalServiceKind | null }[] | undefined
}
export const CommitMessageWithLinks = ({
    message,
    to,
    className,
    onClick,
    externalURLs,
}: Props): React.ReactElement => {
    const commitLinkProps = {
        'data-testid': 'git-commit-node-message-subject',
        className,
        onClick,
        to,
    }

    const github = externalURLs ? externalURLs.find(url => url.serviceKind === ExternalServiceKind.GITHUB) : null
    const matches = [...message.matchAll(GH_ISSUE_NUMBER_IN_COMMIT)]
    if (github && matches.length > 0) {
        const url = githubRepoUrl(github.url)
        let remainingMessage = message
        let skippedCharacters = 0
        const linkSegments: React.ReactNode[] = []

        for (const match of matches) {
            if (match.index === undefined) {
                continue
            }
            const issueNumber = match[2]
            const index = remainingMessage.indexOf(issueNumber, match.index - skippedCharacters)
            const before = remainingMessage.slice(0, index)

            linkSegments.push(
                <Link key={linkSegments.length} {...commitLinkProps}>
                    {before}
                </Link>
            )
            linkSegments.push(
                <Link
                    target="blank"
                    rel="noreferrer noopener"
                    key={linkSegments.length}
                    to={`${url}/pull/${issueNumber.replace('#', '')}`}
                >
                    {issueNumber}
                </Link>
            )

            const nextIndex = index + issueNumber.length
            remainingMessage = remainingMessage.slice(index + issueNumber.length)
            skippedCharacters += nextIndex
        }

        linkSegments.push(
            <Link key={linkSegments.length} {...commitLinkProps}>
                {remainingMessage}
            </Link>
        )

        return <>{linkSegments}</>
    }

    return <Link {...commitLinkProps}>{message}</Link>
}

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
