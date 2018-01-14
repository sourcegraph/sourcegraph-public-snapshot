import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import * as H from 'history'
import * as React from 'react'
import { parseBrowserRepoURL, ParsedRepoURI } from '..'
import { eventLogger } from '../../tracking/eventLogger'

const githubHosts: { [repoPathPrefix: string]: string } = {
    ...(window.context.githubEnterpriseURLs || {}),
    'github.com': 'https://github.com',
}

/**
 * A repository header action that goes to the corresponding URL on GitHub.
 */
export const GoToGitHubAction: React.SFC<{ location: H.Location }> = ({ location }) => {
    const { repoPath, filePath, rev, position, range } = parseBrowserRepoURL(
        location.pathname + location.search + location.hash
    )

    const isGitHub = repoPath.split('/')[0] === 'github.com' || githubHosts[repoPath.split('/')[0]]
    if (!isGitHub) {
        return null
    }

    return (
        <a
            className="btn btn-link btn-sm composite-container__header-action"
            onClick={onClick}
            href={urlToGitHub({ repoPath, filePath, rev, position, range })}
            data-tooltip="View on GitHub"
        >
            <GitHubIcon className="icon-inline" />
        </a>
    )
}

function onClick(): void {
    eventLogger.log('OpenInCodeHostClicked')
}

function urlToGitHub({ repoPath, filePath, rev, position, range }: ParsedRepoURI): string {
    if (!rev) {
        rev = 'HEAD'
    }

    const host = repoPath.split('/')[0]
    const repoURL = `${githubHosts[host]}${repoPath.slice(host.length)}`

    const isDirectory = location.pathname.includes('/-/tree') // TODO(sqs): hacky

    if (filePath) {
        if (isDirectory) {
            return `${repoURL}/tree/${rev}/${filePath}`
        }
        const url = new URL(`${repoURL}/blob/${rev}/${filePath}`)
        if (range) {
            url.hash = `#L${range.start.line}-L${range.end.line}`
        } else if (position) {
            url.hash = '#L' + position.line
        }
        return url.href
    }
    return `${repoURL}/tree/${rev}/`
}
