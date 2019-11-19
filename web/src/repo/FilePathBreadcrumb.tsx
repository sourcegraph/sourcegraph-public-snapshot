import * as React from 'react'
import { LinkOrSpan } from '../../../shared/src/components/LinkOrSpan'
import { RepoRev, toPrettyBlobURL } from '../../../shared/src/util/url'
import { toTreeURL } from '../util/url'

interface Props {
    path: string
    partToUrl: (i: number) => string | undefined
    partToClassName?: (i: number) => string
}

/**
 * A breadcrumb where each path component is a separate link. Use this sparingly. Usually having the entire path be
 * a single link target is more usable; in that case, use RepoFileLink.
 */
const Breadcrumb: React.FunctionComponent<Props> = props => {
    const parts = props.path.split('/')
    const spans: JSX.Element[] = []
    for (const [i, part] of parts.entries()) {
        const link = props.partToUrl(i)
        const className = `part ${props.partToClassName ? props.partToClassName(i) : ''} ${
            i === parts.length - 1 ? 'part-last' : ''
        }`
        spans.push(
            <LinkOrSpan key={i} className={className} to={link}>
                {part}
            </LinkOrSpan>
        )
        if (i < parts.length - 1) {
            spans.push(
                <span key={'sep' + i} className="breadcrumb__separator">
                    /
                </span>
            )
        }
    }
    return (
        // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
        <span className="breadcrumb">{spans}</span>
    )
}

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumb: React.FunctionComponent<RepoRev & {
    filePath: string
    isDir: boolean
}> = ({ repoName, rev, filePath, isDir }) => {
    const parts = filePath.split('/')
    return (
        /* eslint-disable react/jsx-no-bind */
        <Breadcrumb
            path={filePath}
            partToUrl={i => {
                const partPath = parts.slice(0, i + 1).join('/')
                if (isDir || i < parts.length - 1) {
                    return toTreeURL({ repoName, rev, filePath: partPath })
                }
                return toPrettyBlobURL({ repoName, rev, filePath: partPath })
            }}
            partToClassName={i => (i === parts.length - 1 ? 'part-last' : 'part-directory')}
        />
        /* eslint-enable react/jsx-no-bind */
    )
}
