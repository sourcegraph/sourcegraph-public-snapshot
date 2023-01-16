import React, { useCallback, useEffect } from 'react'

import { mdiClose, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { QueryState, SearchContextProps } from '@sourcegraph/shared/src/search'
import { NoResultsSectionID as SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Link, Icon, H3, Text, H2 } from '@sourcegraph/wildcard'

import { QueryExamples } from '../components/QueryExamples'

import { AnnotatedSearchInput } from './AnnotatedSearchExample'

import styles from './NoResultsPage.module.scss'

interface ContainerProps {
    sectionID?: SectionID
    className?: string
    title: string
    children: React.ReactElement | React.ReactElement[]
    onClose?: (sectionID: SectionID) => void
}

const Container: React.FunctionComponent<React.PropsWithChildren<ContainerProps>> = ({
    sectionID,
    title,
    children,
    onClose,
    className = '',
}) => (
    <div className={classNames(styles.container, className)}>
        <H3 className={styles.title}>
            <span className="flex-1">{title}</span>
            {sectionID && (
                <Button variant="icon" aria-label="Hide Section" onClick={() => onClose?.(sectionID)}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            )}
        </H3>
        <div className={styles.content}>{children}</div>
    </div>
)

interface NoResultsPageProps extends ThemeProps, TelemetryProps, Pick<SearchContextProps, 'searchContextsEnabled'> {
    isSourcegraphDotCom: boolean
    showSearchContext: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string
    showQueryExamples?: boolean
    setQueryState?: (query: QueryState) => void
    selectedSearchContextSpec?: string
}

export const NoResultsPage: React.FunctionComponent<React.PropsWithChildren<NoResultsPageProps>> = ({
    searchContextsEnabled,
    isLightTheme,
    telemetryService,
    isSourcegraphDotCom,
    showSearchContext,
    assetsRoot,
    showQueryExamples,
    setQueryState,
    selectedSearchContextSpec,
}) => {
    const [hiddenSectionIDs, setHiddenSectionIds] = useTemporarySetting('search.hiddenNoResultsSections')

    const onClose = useCallback(
        (sectionID: SectionID) => {
            telemetryService.log('NoResultsPanel', { panelID: sectionID, action: 'closed' })
            setHiddenSectionIds((hiddenSectionIDs = []) =>
                !hiddenSectionIDs.includes(sectionID) ? [...hiddenSectionIDs, sectionID] : hiddenSectionIDs
            )
        },
        [setHiddenSectionIds, telemetryService]
    )

    useEffect(() => {
        telemetryService.logViewEvent('NoResultsPage')
    }, [telemetryService])

    return (
        <div className={styles.root}>
            {showQueryExamples && setQueryState && (
                <>
                    <H2 as={H2}>Search basics</H2>
                    <div className={styles.queryExamplesContainer}>
                        <QueryExamples
                            selectedSearchContextSpec={selectedSearchContextSpec}
                            telemetryService={telemetryService}
                            setQueryState={setQueryState}
                            isSourcegraphDotCom={isSourcegraphDotCom}
                        />
                    </div>
                </>
            )}
            <div className={styles.panels}>
                <div className="flex-1 flex-shrink-past-contents">
                    {!hiddenSectionIDs?.includes(SectionID.SEARCH_BAR) && (
                        <Container sectionID={SectionID.SEARCH_BAR} title="The search bar" onClose={onClose}>
                            <div className={styles.annotatedSearchInput}>
                                <AnnotatedSearchInput showSearchContext={searchContextsEnabled && showSearchContext} />
                            </div>
                        </Container>
                    )}

                    <Container title="More resources">
                        <Text>
                            Check out the learn site, including the cheat sheet for more tips on getting the most from
                            Sourcegraph.
                        </Text>
                        <Text>
                            <Link
                                onClick={() => telemetryService.log('NoResultsMore', { link: 'Learn site' })}
                                target="blank"
                                to="https://learn.sourcegraph.com/"
                            >
                                Sourcegraph Learn <Icon svgPath={mdiOpenInNew} aria-label="Open in a new tab" />
                            </Link>
                            <br />
                            <Link
                                onClick={() => telemetryService.log('NoResultsMore', { link: 'Cheat sheet' })}
                                target="blank"
                                to="https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet"
                            >
                                Sourcegraph cheat sheet <Icon svgPath={mdiOpenInNew} aria-label="Open in a new tab" />
                            </Link>
                        </Text>
                    </Container>

                    {hiddenSectionIDs && hiddenSectionIDs.length > 0 && (
                        <Text>
                            Some help panels are hidden.{' '}
                            <Button
                                className="p-0 border-0 align-baseline"
                                onClick={() => {
                                    telemetryService.log('NoResultsPanel', { action: 'showAll' })
                                    setHiddenSectionIds([])
                                }}
                                variant="link"
                            >
                                Show all panels.
                            </Button>
                        </Text>
                    )}
                </div>
            </div>
        </div>
    )
}
