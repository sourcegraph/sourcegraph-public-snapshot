import React, { useCallback, useState, useEffect } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { useHistory } from 'react-router'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { EditorHint, QueryState } from '@sourcegraph/shared/src/search'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Button, H2, Link, Icon, Tabs, TabList, TabPanels, TabPanel, Tab } from '@sourcegraph/wildcard'

import { CloudCtaBanner } from '../../components/CloudCtaBanner'
import { SearchPatternType } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { exampleQueryColumns } from './QueryExamplesHomepage.constants'
import { useQueryExamples, QueryExamplesSection } from './useQueryExamples'

import styles from './QueryExamplesHomepage.module.scss'

export interface QueryExamplesHomepageProps extends TelemetryProps {
    selectedSearchContextSpec?: string
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
    isSourcegraphDotCom?: boolean
    authenticatedUser?: AuthenticatedUser | null
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
    isSourcegraphDotCom = false,
    authenticatedUser,
}) => {
    const [selectedTip, setSelectedTip] = useState<Tip | null>(null)
    const [selectTipTimeout, setSelectTipTimeout] = useState<NodeJS.Timeout>()
    const [queryExampleTabActive, setQueryExampleTabActive] = useState<boolean>(false)
    const history = useHistory()

    const exampleSyntaxColumns = useQueryExamples(selectedSearchContextSpec ?? 'global', isSourcegraphDotCom)

    const handleTabChange = (selectedTab: number): void => {
        setQueryExampleTabActive(!!selectedTab)
    }

    const onQueryExampleClick = useCallback(
        (id: string | undefined, query: string, slug: string | undefined) => {
            // Run search for dotcom longer query examples
            if (isSourcegraphDotCom && queryExampleTabActive) {
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
            queryExampleTabActive,
            history,
            isSourcegraphDotCom,
        ]
    )

    const [cloudCtaVariant, setCloudCtaVariant] = useState<CloudCtaBanner['variant'] | string>('filled')
    useEffect(() => {
        const searchParams = new URL(window.location.href).searchParams
        const uxParam = searchParams.get('cta')
        const allowedVariants: { [key: string]: string | undefined } = {
            a: 'filled',
            b: 'underlined',
            c: 'outlined',
            d: undefined,
        }

        if (uxParam && Object.keys(allowedVariants).includes(uxParam)) {
            setCloudCtaVariant(allowedVariants[uxParam])
        }
    }, [])

    return (
        <div>
            {isSourcegraphDotCom ? (
                <>
                    <div className="d-table mx-auto">
                        <CloudCtaBanner className="mb-5" variant={cloudCtaVariant}>
                            To search across your private repositories,{' '}
                            <Link
                                to={buildCloudTrialURL(authenticatedUser)}
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={() =>
                                    eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'HomeUnderSearch' })
                                }
                            >
                                try Sourcegraph Cloud
                            </Link>
                            .
                        </CloudCtaBanner>
                    </div>

                    <Tabs size="medium" onChange={handleTabChange}>
                        <TabList wrapperClassName={classNames('mb-4', styles.tabHeader)}>
                            <Tab key="Code search basics">Code search basics</Tab>
                            <Tab key="Search query examples">Search query examples</Tab>
                        </TabList>
                        <TabPanels>
                            <TabPanel className={styles.tabPanel}>
                                <QueryExamplesLayout
                                    queryColumns={exampleSyntaxColumns}
                                    onQueryExampleClick={onQueryExampleClick}
                                />
                            </TabPanel>
                            <TabPanel className={styles.tabPanel}>
                                <QueryExamplesLayout
                                    queryColumns={exampleQueryColumns}
                                    onQueryExampleClick={onQueryExampleClick}
                                />
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
                </>
            ) : (
                <div>
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
                    <QueryExamplesLayout
                        queryColumns={exampleSyntaxColumns}
                        onQueryExampleClick={onQueryExampleClick}
                    />
                </div>
            )}
        </div>
    )
}

interface QueryExamplesLayout {
    queryColumns: QueryExamplesSection[][]
    onQueryExampleClick: (id: string | undefined, query: string, slug: string | undefined) => void
}

export const QueryExamplesLayout: React.FunctionComponent<QueryExamplesLayout> = ({
    queryColumns,
    onQueryExampleClick,
}) => (
    <div className={styles.queryExamplesSectionsColumns}>
        {queryColumns.map((column, index) => (
            <div key={`column-${queryColumns[index][0].title}`}>
                {column.map(({ title, queryExamples }) => (
                    <ExamplesSection
                        key={title}
                        title={title}
                        queryExamples={queryExamples}
                        onQueryExampleClick={onQueryExampleClick}
                    />
                ))}
                {/* Add docs link to last column */}
                {queryColumns.length === index + 1 && (
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
)

interface ExamplesSection extends QueryExamplesSection {
    onQueryExampleClick: (id: string | undefined, query: string, slug: string | undefined) => void
}

export const ExamplesSection: React.FunctionComponent<ExamplesSection> = ({
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
    slug?: string | undefined
}

interface QueryExampleChipProps extends QueryExample {
    className?: string
    onClick: (id: string | undefined, query: string, slug: string | undefined) => void | undefined
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
