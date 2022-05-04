import React, { useCallback, useEffect, useRef } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'
import VisibilitySensor from 'react-visibility-sensor'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'
import sanitizeHtml from 'sanitize-html'

import { highlightNode } from '@sourcegraph/common'
import { highlightCode } from '@sourcegraph/search'
import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CommitMatch } from '@sourcegraph/shared/src/search/stream'
import { LoadingSpinner, Link, useEventObservable } from '@sourcegraph/wildcard'

import styles from './CommitSearchResultMatch.module.scss'
import searchResultStyles from './SearchResult.module.scss'

interface CommitSearchResultMatchProps extends PlatformContextProps<'requestGraphQL'> {
    item: CommitMatch
    openInNewTab?: boolean
}

export const CommitSearchResultMatch: React.FunctionComponent<CommitSearchResultMatchProps> = ({
    item,
    openInNewTab,
    platformContext,
}) => {
    const containerElement = useRef<HTMLDivElement>(null)
    const visibilitySensorOffset = { bottom: -500 }

    const getLanguage = useCallback((): string | undefined => {
        const matches = /```(\S+)\s/.exec(item.content)
        if (!matches) {
            return undefined
        }
        return matches[1]
    }, [item.content])

    const [refreshSyntaxHighlighting, syntaxHighlighting] = useEventObservable<string>(
        useCallback(() => {
            const codeContent = item.content.replace(/^```[_a-z]*\n/i, '').replace(/```$/i, '') // Remove Markdown code indicators to render code as plain text

            const lang = getLanguage() || 'txt'

            // Match the code content and any trailing newlines if any.
            if (codeContent) {
                return highlightCode({
                    code: codeContent,
                    fuzzyLanguage: lang,
                    disableTimeout: false,
                    platformContext,
                }).pipe(
                    // Return the rendered markdown if highlighting fails.
                    catchError(error => {
                        console.log(error)
                        return of('<pre>' + sanitizeHtml(item.content) + '</pre>')
                    })
                )
            }

            return of(codeContent)
        }, [getLanguage, item.content, platformContext])
    )

    useEffect((): void => {
        if (containerElement.current) {
            const visibleRows = containerElement.current.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const [line, character, length] of item.ranges) {
                    const code = visibleRows[line - 1]
                    if (code) {
                        highlightNode(code as HTMLElement, character, length)
                    }
                }
            }
        }
    })

    const onChangeVisibility = (isVisible: boolean): void => {
        if (isVisible && typeof syntaxHighlighting !== 'string') {
            refreshSyntaxHighlighting()
        }
    }

    const getFirstLine = (): number => {
        if (item.ranges.length === 0) {
            // If there are no highlights, the calculation below results in -Infinity.
            return 0
        }
        return Math.max(0, Math.min(...item.ranges.map(([line]) => line)) - 1)
    }

    const getLastLine = (): number => {
        if (item.ranges.length === 0) {
            // If there are no highlights, the calculation below results in Infinity,
            // so we set lastLine to 5, which is a just a heuristic for a medium-sized result.
            return 5
        }
        const lastLine = Math.max(...item.ranges.map(([line]) => line)) + 1
        return item.ranges ? Math.min(lastLine, item.ranges.length) : lastLine
    }

    const firstLine = getFirstLine()
    let lastLine = getLastLine()
    if (firstLine === lastLine) {
        // Some edge cases yield the same first and last line, causing the visibility sensor to break, so make sure to avoid
        lastLine++
    }

    const openInNewTabProps = openInNewTab ? { target: '_blank', rel: 'noopener noreferrer' } : undefined

    return (
        <VisibilitySensor
            active={true}
            onChange={onChangeVisibility}
            partialVisibility={true}
            offset={visibilitySensorOffset}
        >
            <div className={styles.commitSearchResultMatch}>
                {item.repoLastFetched && (
                    <LastSyncedIcon className={styles.lastSyncedIcon} lastSyncedTime={item.repoLastFetched} />
                )}
                {syntaxHighlighting !== undefined ? (
                    <Link
                        key={item.url}
                        to={item.url}
                        className={searchResultStyles.searchResultMatch}
                        {...openInNewTabProps}
                    >
                        <code>
                            <Markdown
                                ref={containerElement}
                                testId="search-result-match-code-excerpt"
                                className={classNames(styles.markdown, styles.codeExcerpt)}
                                dangerousInnerHTML={syntaxHighlighting}
                            />
                        </code>
                    </Link>
                ) : (
                    <>
                        <LoadingSpinner className={styles.loader} />
                        <table>
                            <tbody>
                                {range(firstLine, lastLine).map(index => (
                                    <tr key={`${item.url}#${index}`}>
                                        {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                        <td className={styles.lineHidden}>
                                            <code>{index}</code>
                                        </td>
                                        <td className="code"> </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </>
                )}
            </div>
        </VisibilitySensor>
    )
}
