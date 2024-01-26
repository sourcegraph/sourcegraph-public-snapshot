import React, { useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import DOMPurify from 'dompurify'
import { range } from 'lodash'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { highlightNode, logger } from '@sourcegraph/common'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { highlightCode } from '@sourcegraph/shared/src/search'
import type { CommitMatch } from '@sourcegraph/shared/src/search/stream'
import { LoadingSpinner, Link, Code, Markdown } from '@sourcegraph/wildcard'

import styles from './CommitSearchResultMatch.module.scss'
import resultStyles from './ResultContainer.module.scss'

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
    const [highlightedCommitContent, setHighlightedCommitContent] = useState<string>()

    useEffect(() => {
        const langMatches = /```(\S+)\s/.exec(item.content)
        const lang = langMatches ? langMatches[1] : 'txt'

        // Remove Markdown code indicators to render code as plain text
        const codeContent = item.content.replace(/^```[_a-z]*\n/i, '').replace(/```$/i, '')

        const subscription = highlightCode({
            code: codeContent,
            fuzzyLanguage: lang,
            disableTimeout: false,
            platformContext,
        })
            .pipe(
                // Return the rendered markdown if highlighting fails.
                catchError(error => {
                    logger.log(error)
                    return of('<pre>' + DOMPurify.sanitize(item.content) + '</pre>')
                })
            )
            .subscribe(highlightedCommitContent => {
                setHighlightedCommitContent(highlightedCommitContent)
            })

        return () => subscription.unsubscribe()
    }, [item.content, platformContext])

    useEffect((): void => {
        const visibleRows = containerElement.current?.querySelectorAll('table tr')
        if (!visibleRows) {
            return
        }
        for (const [line, character, length] of item.ranges) {
            const code = visibleRows[line - 1]
            if (code) {
                highlightNode(code as HTMLElement, character, length)
            }
        }
    })

    // Subtract 2 to account for the removed leading and trailing ``` code indicators.
    const numLines = useMemo(() => Math.max(1, item.content.split('\n').length - 2), [item.content])

    const openInNewTabProps = openInNewTab ? { target: '_blank', rel: 'noopener noreferrer' } : undefined

    return (
        <Link
            key={item.url}
            to={item.url}
            className={classNames(resultStyles.searchResultMatch, resultStyles.clickable, resultStyles.focusableBlock)}
            {...openInNewTabProps}
        >
            {highlightedCommitContent !== undefined ? (
                <Code>
                    <Markdown
                        ref={containerElement}
                        testId="search-result-match-code-excerpt"
                        className={classNames(styles.markdown, styles.codeExcerpt)}
                        dangerousInnerHTML={highlightedCommitContent}
                    />
                </Code>
            ) : (
                <>
                    <LoadingSpinner className={styles.loader} />
                    <table>
                        <tbody>
                            {range(numLines).map(index => (
                                <tr key={`${item.url}#${index}`}>
                                    {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                    <td className={styles.lineHidden}>
                                        <Code>{index}</Code>
                                    </td>
                                    <td className="code"> </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </>
            )}
        </Link>
    )
}
