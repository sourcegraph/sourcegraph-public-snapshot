import * as React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { toRepoURL, RepoRevision, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { toTreeURL } from '../util/url'

import styles from './FilePathBreadcrumbs.module.scss'

interface Props extends RepoRevision, TelemetryProps {
    filePath: string
    isDir: boolean
}

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumbs: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repoName,
    revision,
    filePath,
    isDir,
    telemetryService,
}) => {
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
            ? 'test-breadcrumb-part-last'
            : classNames('test-breadcrumb-part-directory', styles.partDirectory)

    const spans: JSX.Element[] = [
        <LinkOrSpan
            key="root-dir"
            className={classNames('test-breadcrumb-part-directory', styles.partDirectory)}
            to={toRepoURL({ repoName, revision })}
            aria-current={false}
            onClick={() => telemetryService.log('RootBreadcrumbClicked', { action: 'click', label: 'root directory' })}
        >
            /
        </LinkOrSpan>,
    ]
    for (const [index, part] of parts.entries()) {
        const link = partToUrl(index)
        const className = classNames(styles.part, partToClassName?.(index))
        spans.push(
            <LinkOrSpan
                key={`link-${index}`}
                className={className}
                to={link}
                aria-current={index === parts.length - 1 ? 'page' : 'false'}
            >
                {index < parts.length - 1 ? `${part} /` : part}
            </LinkOrSpan>
        )
    }

    // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
    return <span className={styles.filePathBreadcrumbs}>{spans}</span>
}
