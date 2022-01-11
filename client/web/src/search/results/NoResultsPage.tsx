import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React, { useCallback, useEffect } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button } from '@sourcegraph/wildcard'

import { SearchContextProps } from '..'
import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { useTemporarySetting } from '../../settings/temporary/useTemporarySetting'
import { useExperimentalFeatures } from '../../stores'
import { ModalVideo } from '../documentation/ModalVideo'
import searchBoxStyle from '../input/SearchBox.module.scss'
import searchContextDropDownStyles from '../input/SearchContextDropdown.module.scss'
import { Toggles } from '../input/toggles/Toggles'

import { AnnotatedSearchInput } from './AnnotatedSearchExample'
import styles from './NoResultsPage.module.scss'

export enum SectionID {
    SEARCH_BAR = 'search-bar',
    LITERAL_SEARCH = 'literal-search',
    COMMON_PROBLEMS = 'common-problems',
    VIDEOS = 'videos',
}

const noop = (): void => {}

interface SearchInputExampleProps {
    showSearchContext: boolean
    query: string
    patternType?: SearchPatternType
    runnable?: boolean
    onRun: () => void
}

const SearchInputExample: React.FunctionComponent<SearchInputExampleProps> = ({
    showSearchContext,
    query,
    patternType = SearchPatternType.literal,
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
                        <div className={classNames(searchBoxStyle.searchBoxContextDropdown, styles.fakeSearchContext)}>
                            <div
                                className={classNames(
                                    styles.fakeSearchContextButton,
                                    searchContextDropDownStyles.button,
                                    'btn btn-link text-monospace dropdown-toggle'
                                )}
                            >
                                <code className={searchContextDropDownStyles.buttonContent}>
                                    <span className="search-filter-keyword">context:</span>
                                    global
                                </code>
                            </div>
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
            <Link onClick={onRun} to={{ pathname: '/search', search: builtURLQuery }}>
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

const Container: React.FunctionComponent<ContainerProps> = ({
    sectionID,
    title,
    children,
    onClose,
    className = '',
}) => (
    <div className={classNames(styles.container, className)}>
        <h3 className={styles.title}>
            <span className="flex-1">{title}</span>
            {sectionID && (
                <Button className="btn-icon" aria-label="Hide Section" onClick={() => onClose?.(sectionID)}>
                    <CloseIcon className="icon-inline" />
                </Button>
            )}
        </h3>
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
}

export const NoResultsPage: React.FunctionComponent<NoResultsPageProps> = ({
    searchContextsEnabled,
    isLightTheme,
    telemetryService,
    isSourcegraphDotCom,
}) => {
    const [hiddenSectionIDs, setHiddenSectionIds] = useTemporarySetting('search.hiddenNoResultsSections')
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)

    const onClose = useCallback(
        sectionID => {
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
            <h2>Sourcegraph basics</h2>
            <div className={styles.panels}>
                {!hiddenSectionIDs?.includes(SectionID.VIDEOS) && (
                    <div className="mr-3">
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
                                />
                            ))}
                        </Container>
                    </div>
                )}
                <div className="mr-3 flex-1 flex-shrink-past-contents">
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
                            <p>
                                If you type <code>facebook react</code>, we will search for file names, file contents,
                                repo names, etc. for the exact, ordered phrase <code>facebook react</code>. If you add
                                quotes around your search phrase, we will include the quotes in the search. Literal
                                search makes it easy to find code like:{' '}
                                <code>
                                    {'{'} url: "https://sourcegraph.com" {'}'}
                                </code>{' '}
                                without escaping.
                            </p>
                            <p>
                                Try searching in regexp mode to match terms independently, similar to an AND search, but
                                term ordering is maintained.
                            </p>
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
                            <h4>Finding a specific repository</h4>
                            <p>Repositories are specified by their org/repository-name convention:</p>
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
                            <p>
                                To search within all of an org’s repositories, specify only the org name and a trailing
                                slash:
                            </p>
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
                            <p>
                                <small>
                                    <Link
                                        target="blank"
                                        to="https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet#searching-an-organizations-repository"
                                    >
                                        Learn more <ExternalLinkIcon className="icon-inline" />
                                    </Link>
                                </small>
                            </p>

                            <h4>AND, OR, NOT</h4>
                            <p>Conditionals and grouping are possible within queries:</p>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query="repo:sourcegraph/ (lang:typescript OR lang:go) auth"
                                runnable={isSourcegraphDotCom}
                                onRun={() => telemetryService.log('NoResultsCommonProblems', { search: 'and or' })}
                            />

                            <h4>Escaping</h4>
                            <p>
                                Because our default mode is literal, escaping requires a dedicated filter. Use the
                                content filter to include spaces and filter keywords in searches.
                            </p>
                            <SearchInputExample
                                showSearchContext={searchContextsEnabled && showSearchContext}
                                query={'content:"class Vector"'}
                                runnable={isSourcegraphDotCom}
                                onRun={() => telemetryService.log('NoResultsCommonProblems', { search: 'escaping' })}
                            />
                        </Container>
                    )}

                    <Container title="More resources">
                        <p>
                            Check out the learn site, including the cheat sheet for more tips on getting the most from
                            Sourcegraph.
                        </p>
                        <p>
                            <Link
                                onClick={() => telemetryService.log('NoResultsMore', { link: 'Learn site' })}
                                target="blank"
                                to="https://learn.sourcegraph.com/"
                            >
                                Sourcegraph Learn <ExternalLinkIcon className="icon-inline" />
                            </Link>
                            <br />
                            <Link
                                onClick={() => telemetryService.log('NoResultsMore', { link: 'Cheat sheet' })}
                                target="blank"
                                to="https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet"
                            >
                                Sourcegraph cheat sheet <ExternalLinkIcon className="icon-inline" />
                            </Link>
                        </p>
                    </Container>

                    {hiddenSectionIDs && hiddenSectionIDs.length > 0 && (
                        <p>
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
                        </p>
                    )}
                </div>
            </div>
        </div>
    )
}
