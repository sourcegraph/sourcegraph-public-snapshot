import * as React from 'react'
import { LinkOrSpan } from '../../../shared/src/components/LinkOrSpan'
import { RepoRev, toPrettyBlobURL } from '../../../shared/src/util/url'
import { toTreeURL } from '../util/url'
import classNames from 'classnames'

interface Props {
    path: string
    partToUrl: (i: number) => string | undefined
    partToClassName?: (i: number) => string
}

/**
 * A breadcrumb where each path component is a separate link. Use this sparingly. Usually having the entire path be
 * a single link target is more usable; in that case, use RepoFileLink.
 */
const Breadcrumb: React.FunctionComponent<Props> = ({ path, partToUrl, partToClassName }) => {
    const parts = path.split('/')
    const spans: JSX.Element[] = []
    for (const [index, part] of parts.entries()) {
        const link = partToUrl(index)
        const className = classNames('part', partToClassName?.(index))
        spans.push(
            <LinkOrSpan key={index} className={className} to={link}>
                {part}
            </LinkOrSpan>
        )
        if (index < parts.length - 1) {
            spans.push(
                <span key={`sep${index}`} className="breadcrumb__separator">
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
export const FilePathBreadcrumb: React.FunctionComponent<
    RepoRev & {
        filePath: string
        isDir: boolean
    }
> = ({ repoName, revision, filePath, isDir }) => {
    const parts = filePath.split('/')
    return (
        /* eslint-disable react/jsx-no-bind */
        <Breadcrumb
            path={filePath}
            partToUrl={index => {
                const partPath = parts.slice(0, index + 1).join('/')
                if (isDir || index < parts.length - 1) {
                    return toTreeURL({ repoName, revision, filePath: partPath })
                }
                return toPrettyBlobURL({ repoName, revision, filePath: partPath })
            }}
            partToClassName={index =>
                index === parts.length - 1
                    ? 'part-last test-breadcrumb-part-last'
                    : 'part-directory test-breadcrumb-part-directory'
            }
        />
        /* eslint-enable react/jsx-no-bind */
    )
}
