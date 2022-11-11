import React, { useCallback, useEffect } from 'react'

import { mdiClose, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { SearchContextProps } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery, Toggles } from '@sourcegraph/search-ui'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { NoResultsSectionID as SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link, Icon, Code, H2, H3, H4, Text, createLinkUrl } from '@sourcegraph/wildcard'

import { ModalVideo } from '../documentation/ModalVideo'

import { AnnotatedSearchInput } from './AnnotatedSearchExample'

import searchBoxStyle from '../input/SearchBox.module.scss'
import searchContextDropDownStyles from '../input/SearchContextDropdown.module.scss'
import styles from './NoResultsPage.module.scss'

const noop = (): void => {}

interface SearchInputExampleProps {
    showSearchContext: boolean
    query: string
    patternType?: SearchPatternType
    runnable?: boolean
    onRun: () => void
}

const SearchInputExample: React.FunctionComponent<React.PropsWithChildren<SearchInputExampleProps>> = ({
    showSearchContext,
    query,
    patternType = SearchPatternType.standard,
    runnable = false,
    onRun,
}) => {
    const example = (
        <div className={classNames(searchBoxStyle.searchBox, styles.fakeSearchbox)}>
            <div
                className={classNames(
                    searchBoxStyle.searchBoxBackgroundContainer,
                    styles.fakeSearchboxBackgroundContainer,
                    'flex-shrink-past-contents'
                )}
            >
                {showSearchContext && (
                    <>
                        <div className={classNames(searchBoxStyle.searchBoxAlignMiddle, styles.fakeSearchContext)}>
                            <Button
                                as="div"
                                variant="link"
                                className={classNames(
                                    styles.fakeSearchContextButton,
                                    searchContextDropDownStyles.button,
                                    'text-monospace dropdown-toggle'
                                )}
                            >
                                <Code className={searchContextDropDownStyles.buttonContent}>
                                    {/*
                                       a11y-ignore
                                       Rule: "color-contrast" (Elements must have sufficient color contrast)
                                       GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                                     */}
                                    <span className="search-filter-keyword a11y-ignore">context:</span>
                                    global
                                </Code>
                            </Button>
                        </div>
                        <div className={classNames(searchBoxStyle.searchBoxSeparator, styles.fakeSearchboxSeparator)} />
                    </>
                )}
                <div
                    className={classNames(
                        searchBoxStyle.searchBoxFocusContainer,
                        styles.fakeSearchboxFocusContainer,
                        'flex-shrink-past-contents'
                    )}
                >
                    <div
                        className={classNames(
                            searchBoxStyle.searchBoxInput,
                            styles.fakeSearchInput,
                            'flex-shrink-past-contents'
                        )}
                    >
                        <SyntaxHighlightedSearchQuery query={query} />
                    </div>
                </div>
                <div className={styles.fakeSearchboxToggles}>
                    <Toggles
                        navbarSearchQuery={query}
                        caseSensitive={false}
                        patternType={patternType}
                        setCaseSensitivity={noop}
                        setPatternType={noop}
                        settingsCascade={{ subjects: null, final: {} }}
                        showCopyQueryButton={false}
                        interactive={false}
                    />
                </div>
            </div>
        </div>
    )

    if (runnable) {
        const builtURLQuery = buildSearchURLQuery(query, patternType, false, 'global')
        return (
            <Link onClick={onRun} to={createLinkUrl({ pathname: '/search', search: builtURLQuery })}>
                <div className={styles.searchInputExample}>
                    {example}
                    <span className="ml-2 text-nowrap">Run Search</span>
                </div>
            </Link>
        )
    }
    return <div className={styles.searchInputExample}>{example}</div>
}

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

const videos = [
    {
        title: 'Three ways to search',
        thumbnailPrefix: 'img/vt-three-ways-to-search',
        src: 'https://www.youtube-nocookie.com/embed/XLfE2YuRwvw',
    },
    {
        title: 'Diff and commit search',
        thumbnailPrefix: 'img/vt-diff-and-commit-search',
        src: 'https://www.youtube-nocookie.com/embed/w-RrDz9hyGI',
    },
    {
        title: 'Finding error messages',
        thumbnailPrefix: 'img/vt-finding-error-messages',
        src: 'https://www.youtube-nocookie.com/embed/r2CpLe1h89I',
    },
    {
        title: 'Structural search',
        thumbnailPrefix: 'img/vt-structural-search',
        src: 'https://www.youtube-nocookie.com/embed/GnubTdnilbc',
    },
]

interface NoResultsPageProps extends ThemeProps, TelemetryProps, Pick<SearchContextProps, 'searchContextsEnabled'> {
    isSourcegraphDotCom: boolean
    showSearchContext: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string
}

export const NoResultsPage: React.FunctionComponent<React.PropsWithChildren<NoResultsPageProps>> = ({
    searchContextsEnabled,
    isLightTheme,
    telemetryService,
    isSourcegraphDotCom,
    showSearchContext,
    assetsRoot,
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
            <H2>Sourcegraph basics</H2>
            <div className={styles.panels}>
                {!hiddenSectionIDs?.includes(SectionID.VIDEOS) && (
                    <Container
                        sectionID={SectionID.VIDEOS}
                        title="Video explanations"
                        className={styles.videoContainer}
                        onClose={onClose}
                    >
                        {videos.map(video => (
                            <ModalVideo
                                key={video.title}
                                className={styles.video}
                                id={`video-${video.title.toLowerCase().replace(/[^a-z]+/, '-')}`}
                                title={video.title}
                                src={video.src}
                                thumbnail={{
                                    src: `${video.thumbnailPrefix}-${isLightTheme ? 'light' : 'dark'}.png`,
                                    alt: `${video.title} video thumbnail`,
                                }}
                                showCaption={true}
                                onToggle={isOpen => {
                                    if (isOpen) {
                                        telemetryService.log('NoResultsVideoPlayed', { video: video.title })
                                    }
                                }}
                                assetsRoot={assetsRoot}
                            />
                        ))}
                    </Container>
                )}
                <div className="flex-1 flex-shrink-past-contents">
                    {!hiddenSectionIDs?.includes(SectionID.SEARCH_BAR) && (
                        <Container sectionID={SectionID.SEARCH_BAR} title="The search bar" onClose={onClose}>
                            <div className={styles.annotatedSearchInput}>
                                <AnnotatedSearchInput showSearchContext={searchContextsEnabled && showSearchContext} />
                            </div>
                        </Container>
                    )}
                    {!hiddenSectionIDs?.includes(SectionID.LITERAL_SEARCH) && (
                        <Container
                            sectionID={SectionID.LITERAL_SEARCH}
                            title="Search is literal by default"
                            onClose={onClose}
                        >
                            <Text>
                                If you type <Code>facebook react</Code>, we will search for file names, file contents,
                                repo names, etc. for the exact, ordered phrase <Code>facebook react</Code>. If you add
                                quotes around your search phrase, we will include the quotes in the search. Literal
                                search makes it easy to find code like:{' '}
                                <Code>
                                    {'{'} url: "https://sourcegraph.com" {'}'}
                                </Code>{' '}
                                without escaping.
                            </Text>
                            <Text>
                                Try searching in regexp mode to match terms independently, similar to an AND search, but
                                term ordering is maintained.
                            </Text>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query="repo:sourcegraph const Authentication"
                                patternType={SearchPatternType.regexp}
                                runnable={isSourcegraphDotCom}
                                onRun={() =>
                                    telemetryService.log('NoResultsSearchLiteral', { search: 'regexp search' })
                                }
                            />
                        </Container>
                    )}
                    {!hiddenSectionIDs?.includes(SectionID.COMMON_PROBLEMS) && (
                        <Container sectionID={SectionID.COMMON_PROBLEMS} title="Common Problems" onClose={onClose}>
                            <H4>Finding a specific repository</H4>
                            <Text>Repositories are specified by their org/repository-name convention:</Text>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query="repo:sourcegraph/about lang:go publish"
                                runnable={isSourcegraphDotCom}
                                onRun={() =>
                                    telemetryService.log('NoResultsCommonProblems', {
                                        search: 'zfind specific repo',
                                    })
                                }
                            />
                            <Text>
                                To search within all of an org’s repositories, specify only the org name and a trailing
                                slash:
                            </Text>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query="repo:sourcegraph/ lang:go publish"
                                runnable={isSourcegraphDotCom}
                                onRun={() =>
                                    telemetryService.log('NoResultsCommonProblems', {
                                        search: 'find specific repo',
                                    })
                                }
                            />
                            <Text>
                                <small>
                                    <Link
                                        target="blank"
                                        to="https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet#searching-an-organizations-repository"
                                    >
                                        Learn more <Icon aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                                    </Link>
                                </small>
                            </Text>

                            <H4>AND, OR, NOT</H4>
                            <Text>Conditionals and grouping are possible within queries:</Text>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query="repo:sourcegraph/ (lang:typescript OR lang:go) auth"
                                runnable={isSourcegraphDotCom}
                                onRun={() => telemetryService.log('NoResultsCommonProblems', { search: 'and or' })}
                            />

                            <H4>Escaping</H4>
                            <Text>
                                Because our default mode is literal, escaping requires a dedicated filter. Use the
                                content filter to include spaces and filter keywords in searches.
                            </Text>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query={'content:"class Vector"'}
                                runnable={isSourcegraphDotCom}
                                onRun={() => telemetryService.log('NoResultsCommonProblems', { search: 'escaping' })}
                            />
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
