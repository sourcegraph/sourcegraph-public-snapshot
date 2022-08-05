import React, { useCallback, useState } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { EditorHint, QueryState } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Link } from '@sourcegraph/wildcard'

import styles from './QueryExamplesHomepage.module.scss'

export interface QueryExamplesHomepageProps extends TelemetryProps {
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
}

const queryExampleSectionsColumns = [
    [
        {
            title: 'Scope search to specific repos',
            queryExamples: [
                { id: 'single-repo', query: 'repo:awesomecorp/big-repo' },
                { id: 'org-repos', query: 'repo:awesomecorp/*' },
            ],
        },
        {
            title: 'Jump into code navigation',
            queryExamples: [
                { id: 'file-filter', query: 'file:examplefile.go' },
                { id: 'type-symbol', query: 'type:symbol Handler' },
            ],
        },
        {
            title: 'Get specific',
            queryExamples: [
                { id: 'author', query: 'author:logansmith' },
                { id: 'before-after-filters', query: 'before:today after:earlydate' },
            ],
        },
    ],
    [
        {
            title: 'Find exact matches',
            queryExamples: [{ id: 'exact-matches', query: 'some error message', helperText: 'No quotes needed' }],
        },
        {
            title: 'Operators',
            queryExamples: [
                { id: 'or-operator', query: 'lang:javascript OR lang:typescript' },
                { id: 'and-operator', query: 'example AND secondexample' },
                { id: 'not-operator', query: 'lang:go NOT file:main.go' },
            ],
        },
        {
            title: 'Get advanced',
            queryExamples: [{ id: 'contains-commit-after', query: 'repo:contains.commit.after(yesterday)' }],
            footer: (
                <small className="d-block mt-3">
                    <Link target="blank" to="/help/code_search/reference/queries">
                        Complete query reference{' '}
                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                    </Link>
                </small>
            ),
        },
    ],
]

type Tip = 'rev' | 'lang' | 'type-commit-diff'

export const queryToTip = (id: string | undefined): Tip | null => {
    switch (id) {
        case 'single-repo':
        case 'org-repos':
            return 'rev'
        case 'file-filter':
        case 'type-symbol':
        case 'exact-matches':
            return 'lang'
        case 'author':
            return 'type-commit-diff'
    }
    return null
}

export const QueryExamplesHomepage: React.FunctionComponent<QueryExamplesHomepageProps> = ({
    telemetryService,
    queryState,
    setQueryState,
}) => {
    const [selectedTip, setSelectedTip] = useState<Tip | null>(null)
    const [selectTipTimeout, setSelectTipTimeout] = useState<NodeJS.Timeout>()

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
                {selectedTip === 'type-commit-diff' && (
                    <>
                        Use <QueryExampleChip query="type:commit" onClick={onQueryExampleClick} className="mx-1" /> or{' '}
                        <QueryExampleChip query="type:diff" onClick={onQueryExampleClick} className="mx-1" /> to specify
                        where the author appears
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
        <div className={styles.queryExamplesSectionTitle}>{title}</div>
        <div className={styles.queryExamplesItems}>
            {queryExamples.map(({ id, query, helperText }) => (
                <QueryExampleChip
                    id={id}
                    key={query}
                    query={query}
                    helperText={helperText}
                    onClick={onQueryExampleClick}
                />
            ))}
        </div>
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
    <span className={classNames('d-flex align-items-center', className)}>
        <Button type="button" className={styles.queryExampleChip} onClick={() => onClick(id, query)}>
            <SyntaxHighlightedSearchQuery query={query} />
        </Button>
        {helperText && (
            <span className="text-muted ml-2">
                <small>{helperText}</small>
            </span>
        )}
    </span>
)
