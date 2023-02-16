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
        rel: 'noreferrer noopener',
        target: '_blank',
        to,
    }

    const github = externalURLs ? externalURLs.find(url => url.serviceKind === ExternalServiceKind.GITHUB) : null
    const matches = [...message.matchAll(GH_ISSUE_NUMBER_IN_COMMIT)]
    if (github && matches.length > 0) {
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

            linkSegments.push(<Link {...commitLinkProps}>{before}</Link>)
            linkSegments.push(<Link to={`${github.url}/pull/${issueNumber.replace('#', '')}`}>{issueNumber}</Link>)

            const nextIndex = index + issueNumber.length
            remainingMessage = remainingMessage.slice(index + issueNumber.length)
            skippedCharacters += nextIndex
        }

        linkSegments.push(<Link {...commitLinkProps}>{remainingMessage}</Link>)

        return <>{linkSegments}</>
    }

    return <Link {...commitLinkProps}>{message}</Link>
}
