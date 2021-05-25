import classNames from 'classnames'
import * as React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { RepoRevision, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { toTreeURL } from '../util/url'

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumbs: React.FunctionComponent<
    RepoRevision & {
        filePath: string
        isDir: boolean
    }
> = ({ repoName, revision, filePath, isDir }) => {
    const parts = filePath.split('/')
    const partToUrl = (index: number): string => {
        const partPath = parts.slice(0, index + 1).join('/')
        if (isDir || index < parts.length - 1) {
            return toTreeURL({ repoName, revision, filePath: partPath })
        }
        return toPrettyBlobURL({ repoName, revision, filePath: partPath })
    }
    const partToClassName = (index: number): string =>
        index === parts.length - 1
            ? 'font-weight-bold test-breadcrumb-part-last'
            : 'part-directory test-breadcrumb-part-directory'

    const spans: JSX.Element[] = []
    for (const [index, part] of parts.entries()) {
        const link = partToUrl(index)
        const className = classNames('part', partToClassName?.(index))
        spans.push(
            <LinkOrSpan
                key={index}
                className={className}
                to={link}
                aria-current={index === parts.length - 1 ? 'page' : 'false'}
            >
                {part}
            </LinkOrSpan>
        )
        if (index < parts.length - 1) {
            spans.push(
                <span key={`sep${index}`} className="file-path-breadcrumbs__separator text-muted font-weight-medium">
                    /
                </span>
            )
        }
    }
    return (
        // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
        <span className="file-path-breadcrumbs">
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory-2
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory-3
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory-4
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory-5
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-directory test-breadcrumb-part-directory" to="/">
                Directory-6
            </LinkOrSpan>
            <span className="file-path-breadcrumbs__separator text-muted font-weight-medium">/</span>
            <LinkOrSpan className="part part-file font-weight-bold test-breadcrumb-part-last" to="/">
                Deeply-nested-file-name-that-is-long-and-hard-to-find.tsx
            </LinkOrSpan>
        </span>
    )
}
