import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { EditorHint, QueryState, SearchPatternType } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, H2 } from '@sourcegraph/wildcard'

import { useQueryExamples } from './useQueryExamples'

import styles from './QueryExamplesHomepage.module.scss'

export interface QueryExamplesHomepageProps extends TelemetryProps {
    selectedSearchContextSpec?: string
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
}

type Tip = 'rev' | 'lang' | 'before'

export const queryToTip = (id: string | undefined): Tip | null => {
    switch (id) {
        case 'single-repo':
        case 'org-repos':
            return 'rev'
        case 'exact-matches':
        case 'regex-pattern':
            return 'lang'
        case 'type-diff-author':
        case 'type-commit-message':
        case 'type-diff-after':
            return 'before'
    }
    return null
}

export const QueryExamplesHomepage: React.FunctionComponent<QueryExamplesHomepageProps> = ({
    selectedSearchContextSpec,
    telemetryService,
    queryState,
    setQueryState,
}) => {
    const [selectedTip, setSelectedTip] = useState<Tip | null>(null)
    const [selectTipTimeout, setSelectTipTimeout] = useState<NodeJS.Timeout>()

    const queryExampleSectionsColumns = useQueryExamples(selectedSearchContextSpec ?? 'global')

    const onQueryExampleClick = useCallback(
        (id: string | undefined, query: string) => {
            setQueryState({ query: `${queryState.query} ${query}`.trimStart(), hint: EditorHint.Focus })

            telemetryService.log('QueryExampleClicked', { queryExample: query }, { queryExample: query })

            // Clear any previously set timeout.
            if (selectTipTimeout) {
                clearTimeout(selectTipTimeout)
            }

            const newSelectedTip = queryToTip(id)
            if (newSelectedTip) {
                // If the user selected a query with a different tip, reset the currently selected tip, so that we
                // can apply the fade-in transition.
                if (newSelectedTip !== selectedTip) {
                    setSelectedTip(null)
                }

                const timeoutId = setTimeout(() => setSelectedTip(newSelectedTip), 1000)
                setSelectTipTimeout(timeoutId)
            } else {
                // Immediately reset the selected tip if the query does not have an associated tip.
                setSelectedTip(null)
            }
        },
        [
            telemetryService,
            queryState.query,
            setQueryState,
            selectedTip,
            setSelectedTip,
            selectTipTimeout,
            setSelectTipTimeout,
        ]
    )

    return (
        <div>
            <div className={classNames(styles.tip, selectedTip && styles.tipVisible)}>
                <strong>Tip</strong>
                <span className="mx-1">–</span>
                {selectedTip === 'rev' && (
                    <>
                        Add <QueryExampleChip query="rev:branchname" onClick={onQueryExampleClick} className="mx-1" />{' '}
                        to query accross a specific branch or commit
                    </>
                )}
                {selectedTip === 'lang' && (
                    <>
                        Use <QueryExampleChip query="lang:" onClick={onQueryExampleClick} className="mx-1" /> to query
                        for matches only in a given language
                    </>
                )}
                {selectedTip === 'before' && (
                    <>
                        Use{' '}
                        <QueryExampleChip query={'before:"last week"'} onClick={onQueryExampleClick} className="mx-1" />{' '}
                        to query within a time range
                    </>
                )}
            </div>
            <div className={styles.queryExamplesSectionsColumns}>
                {queryExampleSectionsColumns.map((column, index) => (
                    // eslint-disable-next-line react/no-array-index-key
                    <div key={`column-${index}`}>
                        {column.map(({ title, queryExamples, footer }) => (
                            <QueryExamplesSection
                                key={title}
                                title={title}
                                queryExamples={queryExamples}
                                footer={footer}
                                onQueryExampleClick={onQueryExampleClick}
                            />
                        ))}
                    </div>
                ))}
            </div>
        </div>
    )
}

interface QueryExamplesSection {
    title: string
    queryExamples: QueryExample[]
    footer?: React.ReactElement
    onQueryExampleClick: (id: string | undefined, query: string) => void
}

export const QueryExamplesSection: React.FunctionComponent<QueryExamplesSection> = ({
    title,
    queryExamples,
    footer,
    onQueryExampleClick,
}) => (
    <div className={styles.queryExamplesSection}>
        <H2 className={styles.queryExamplesSectionTitle}>{title}</H2>
        <ul className={classNames('list-unstyled', styles.queryExamplesItems)}>
            {queryExamples
                .filter(({ query }) => query.length > 0)
                .map(({ id, query, helperText }) => (
                    <QueryExampleChip
                        id={id}
                        key={query}
                        query={query}
                        helperText={helperText}
                        onClick={onQueryExampleClick}
                    />
                ))}
        </ul>
        {footer}
    </div>
)

interface QueryExample {
    id?: string
    query: string
    helperText?: string
}

interface QueryExampleChipProps extends QueryExample {
    className?: string
    onClick: (id: string | undefined, query: string) => void
}

export const QueryExampleChip: React.FunctionComponent<QueryExampleChipProps> = ({
    id,
    query,
    helperText,
    className,
    onClick,
}) => (
    <li className={classNames('d-flex align-items-center', className)}>
        <Button type="button" className={styles.queryExampleChip} onClick={() => onClick(id, query)}>
            <SyntaxHighlightedSearchQuery query={query} searchPatternType={SearchPatternType.standard} />
        </Button>
        {helperText && (
            <span className="text-muted ml-2">
                <small>{helperText}</small>
            </span>
        )}
    </li>
)
