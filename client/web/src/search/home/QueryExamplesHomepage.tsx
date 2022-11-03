import React, { useCallback, useState } from 'react'

import { useHistory } from 'react-router'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { EditorHint, QueryState, SearchPatternType } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, H2, H4, Text, Link, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'

import { useQueryExamples } from './useQueryExamples'

import styles from './QueryExamplesHomepage.module.scss'

export interface QueryExamplesHomepageProps extends TelemetryProps {
    selectedSearchContextSpec?: string
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
    isSourcegraphDotCom?: boolean
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

const exampleQueryColumns = [
    [
        {
            title: 'Discover how developers are using hooks',
            queryExamples: [{ id: 'hooks', query: 'useState AND useRef lang:javascript', slug: '?q=context:global+useState+AND+useRef+lang:javascript' }],
        },
        {
            title: 'Discover what is being passed to propTypes for type checking',
            queryExamples: [{ id: 'prop-types', query: '.propTypes = {...} patterntype:structural', slug: '?q=context:global+.propTypes+%3D+%7B...%7D' }],
        },
    ],
    [
        {
            title: 'Error boundaries in React',
            queryExamples: [{ id: 'error-boundaries', query: 'static getDerivedStateFromError(', slug: '' }],
        },
        {
            title: 'Find to-dos on a specific repository',
            queryExamples: [{ id: 'find-todos', query: 'repo:facebook/react content:TODO', slug: '' }],
        },
    ],
]

export const QueryExamplesHomepage: React.FunctionComponent<QueryExamplesHomepageProps> = ({
    selectedSearchContextSpec,
    telemetryService,
    queryState,
    setQueryState,
    isSourcegraphDotCom = false,
}) => {
    const [selectedTip, setSelectedTip] = useState<Tip | null>(null)
    const [selectTipTimeout, setSelectTipTimeout] = useState<NodeJS.Timeout>()
    const [toggleSyntaxToQueryExamples, setToggleSyntaxToQueryExamples] = useState<boolean>(false)
    const history = useHistory()

    const exampleSyntaxColumns = useQueryExamples(selectedSearchContextSpec ?? 'global', isSourcegraphDotCom)

    const onQueryExampleClick = useCallback(
        (id: string | undefined, query: string, slug: string | undefined) => {
            console.log('hit', isSourcegraphDotCom, toggleSyntaxToQueryExamples)
            if (isSourcegraphDotCom && toggleSyntaxToQueryExamples) {
                telemetryService.log('QueryExampleClicked', { queryExample: query }, { queryExample: query })
                history.push(slug!)
            }
            
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
            <div>
                {isSourcegraphDotCom ? (
                    <div className="d-flex justify-content-center align-items-center mb-2">
                        <Text className={classNames('mr-2 pr-2', styles.codeBasicsTitle)}>
                            {toggleSyntaxToQueryExamples ? 'Search Query Examples' : 'Code Search Basics'}
                        </Text>
                        <Button onClick={() => setToggleSyntaxToQueryExamples(!toggleSyntaxToQueryExamples)}>
                            {toggleSyntaxToQueryExamples ? 'Code search basics' : 'Search query examples'}
                        </Button>
                    </div>
                ) : (
                    <div className={classNames(styles.tip, selectedTip && styles.tipVisible)}>
                        <strong>Tip</strong>
                        <span className="mx-1">–</span>
                        {selectedTip === 'rev' && (
                            <>
                                Add{' '}
                                <QueryExampleChip
                                    query="rev:branchname"
                                    onClick={onQueryExampleClick}
                                    className="mx-1"
                                />{' '}
                                to query accross a specific branch or commit
                            </>
                        )}
                        {selectedTip === 'lang' && (
                            <>
                                Use <QueryExampleChip query="lang:" onClick={onQueryExampleClick} className="mx-1" /> to
                                query for matches only in a given language
                            </>
                        )}
                        {selectedTip === 'before' && (
                            <>
                                Use{' '}
                                <QueryExampleChip
                                    query={'before:"last week"'}
                                    onClick={onQueryExampleClick}
                                    className="mx-1"
                                />{' '}
                                to query within a time range
                            </>
                        )}
                    </div>
                )}
                <div className={styles.queryExamplesSectionsColumns}>
                    {(isSourcegraphDotCom && toggleSyntaxToQueryExamples ? exampleQueryColumns : exampleSyntaxColumns).map((column, index) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <div key={`column-${index}`}>
                            {column.map(({ title, queryExamples }) => (
                                <QueryExamplesSection
                                    key={title}
                                    title={title}
                                    queryExamples={queryExamples}
                                    onQueryExampleClick={onQueryExampleClick}
                                />
                            ))}
                            {!!index && (
                                <small className="d-block">
                                    <Link target="blank" to="/help/code_search/reference/queries">
                                        Complete query reference{' '}
                                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                                    </Link>
                                </small>
                            )}
                        </div>
                    ))}
                </div>
            </div>
            {isSourcegraphDotCom && (
                <div className="d-flex align-items-center justify-content-lg-center my-5">
                    <H4 className={classNames('mr-2 mb-0 pr-2', styles.proTipTitle)}>Pro Tip</H4>
                    <Link to="https://signup.sourcegraph.com/" onClick={() => eventLogger.log('ClickedOnCloudCTA')}>
                        Use Sourcegraph to search across your team's code.
                    </Link>
                </div>
            )}
        </div>
    )
}

interface QueryExamplesSection {
    title: string
    queryExamples: QueryExample[]
    onQueryExampleClick: (id: string | undefined, query: string, slug: string | undefined) => void
}

export const QueryExamplesSection: React.FunctionComponent<QueryExamplesSection> = ({
    title,
    queryExamples,
    onQueryExampleClick,
}) => (
    <div className={styles.queryExamplesSection}>
        <H2 className={styles.queryExamplesSectionTitle}>{title}</H2>
        <ul className={classNames('list-unstyled', styles.queryExamplesItems)}>
            {queryExamples
                .filter(({ query }) => query.length > 0)
                .map(({ id, query, helperText, slug }) => (
                    <QueryExampleChip
                        id={id}
                        key={query}
                        query={query}
                        slug={slug}
                        helperText={helperText}
                        onClick={onQueryExampleClick}
                    />
                ))}
        </ul>
    </div>
)

interface QueryExample {
    id?: string
    query: string
    helperText?: string
    slug?: string
}

interface QueryExampleChipProps extends QueryExample {
    className?: string
    onClick: (id: string | undefined, query: string, slug?: string | undefined) => void
}

export const QueryExampleChip: React.FunctionComponent<QueryExampleChipProps> = ({
    id,
    query,
    helperText,
    slug,
    className,
    onClick,
}) => (
    <li className={classNames('d-flex align-items-center', className)}>
        <Button type="button" className={styles.queryExampleChip} onClick={() => onClick(id, query, slug || '')}>
            <SyntaxHighlightedSearchQuery query={query} searchPatternType={SearchPatternType.standard} />
        </Button>
        {helperText && (
            <span className="text-muted ml-2">
                <small>{helperText}</small>
            </span>
        )}
    </li>
)
