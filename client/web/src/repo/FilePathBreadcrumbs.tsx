import * as React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { toRepoURL, RepoRevision, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { toTreeURL } from '../util/url'

import styles from './FilePathBreadcrumbs.module.scss'

interface Props extends RepoRevision, TelemetryProps {
    filePath: string
    isDir: boolean
}

const MAXIMUM_DIRECTORIES = 8

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

    const shouldTruncate = parts.length > MAXIMUM_DIRECTORIES

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
    if (shouldTruncate) {
        const truncatedParts = parts.slice(0, parts.length - MAXIMUM_DIRECTORIES)
        const truncatedPath = truncatedParts.join('/')
        const link = partToUrl(truncatedParts.length - 1)
        spans.push(
            <FilePath
                key="truncated-dir"
                className={classNames('test-breadcrumb-part-truncated', styles.partDirectory)}
                link={link}
                isLast={false}
                label="..."
                fullLabel={truncatedPath}
            />
        )
    }

    for (const [index, part] of parts.entries()) {
        if (shouldTruncate && index < parts.length - MAXIMUM_DIRECTORIES) {
            continue
        }

        const link = partToUrl(index)
        const className = classNames(styles.part, partToClassName?.(index))
        const isLast = index === parts.length - 1
        spans.push(
            <FilePath
                key={`link-${index}`}
                className={className}
                link={link}
                isLast={index === parts.length - 1}
                label={part}
            />
        )
        if (!isLast) {
            spans.push()
        }
    }

    // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
    return <span className={styles.filePathBreadcrumbs}>{spans}</span>
}

interface FilePathProps {
    label: string
    isLast: boolean
    className: string
    link: string
    fullLabel?: string
}
function FilePath({ label, isLast, className, link, fullLabel }: FilePathProps): JSX.Element {
    const [ref, truncated, checkTruncation] = useIsTruncated<HTMLAnchorElement>()
    return (
        <>
            <Tooltip content={fullLabel || (truncated ? label : null)}>
                <LinkOrSpan
                    className={className}
                    to={link}
                    onFocus={checkTruncation}
                    onMouseEnter={checkTruncation}
                    aria-current={isLast ? 'page' : 'false'}
                    aria-label={fullLabel}
                    ref={ref}
                >
                    {label}
                </LinkOrSpan>
            </Tooltip>
            {!isLast ? <span className={styles.partSeparator}>/</span> : null}
        </>
    )
}
