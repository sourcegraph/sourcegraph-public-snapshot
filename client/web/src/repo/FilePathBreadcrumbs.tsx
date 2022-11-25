import * as React from 'react'
import { useLayoutEffect } from 'react'

import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

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

    const [truncatedElements, setTruncatedElements] = React.useState<number>(0)

    // Increase the number of truncatedElements and verify if the container is
    // still overflowing, up to the point where only the the current element is
    // visible.
    //
    // Warning: This might cause a few re-renders of the file breadcrumbs until
    // a fitting number of truncated elements is found. However the changes
    // necessary to do this are bound by the number of folders in the file path.
    const ref = React.useRef<HTMLDivElement>(null)
    const { width } = useResizeObserver({ ref })
    // Reset the truncation logic when the element is resized
    useLayoutEffect(() => setTruncatedElements(0), [width])
    useLayoutEffect(() => {
        const element = ref.current
        if (!element || truncatedElements >= parts.length - 1) {
            return
        }
        const isOverflowing = Math.ceil(element.scrollWidth - element.clientWidth - element.scrollLeft) > 0
        if (isOverflowing) {
            setTruncatedElements(truncatedElements + 1)
        }
    }, [parts.length, truncatedElements])

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

    if (truncatedElements > 0) {
        const truncatedParts = parts.slice(0, truncatedElements)
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
        if (index < truncatedElements) {
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
    return (
        <span ref={ref} className={styles.filePathBreadcrumbs}>
            {spans}
        </span>
    )
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
