import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { useTemporarySetting } from '../../settings/temporary/useTemporarySetting'
import { ModalVideo } from '../documentation/ModalVideo'
import searchBoxStyle from '../input/SearchBox.module.scss'
import { Toggles } from '../input/toggles/Toggles'

import styles from './NoResultsPage.module.scss'

export enum SectionID {
    SEARCH_BAR = 'search-bar',
    LITERAL_SEARCH = 'literal-search',
    COMMON_PROBLEMS = 'common-problems',
    VIDEOS = 'videos',
}

const SearchContext: React.FunctionComponent<{}> = () => (
    <>
        <div className="btn text-monospace">
            <code>
                <span className="search-filter-keyword">context:</span>
                global
            </code>
        </div>
        <div className={searchBoxStyle.searchBoxSeparator} />
    </>
)

const noop = (): void => {}

interface SearchInputExampleProps {
    showSearchContext: boolean
    query: string
    patternType?: SearchPatternType
}

const SearchInputExample: React.FunctionComponent<SearchInputExampleProps> = ({
    showSearchContext,
    query,
    patternType = SearchPatternType.literal,
}) => (
    <p>
        <div
            className={classNames(
                searchBoxStyle.searchBox,
                searchBoxStyle.searchBoxBackgroundContainer,
                styles.searchBox
            )}
        >
            {showSearchContext && <SearchContext />}
            <span className={searchBoxStyle.searchBoxFocusContainer}>
                <SyntaxHighlightedSearchQuery query={query} />
            </span>
            <Toggles
                navbarSearchQuery={query}
                caseSensitive={false}
                history={null}
                location={null}
                patternType={patternType}
                setCaseSensitivity={noop}
                setPatternType={noop}
                settingsCascade={{ subjects: null, final: {} }}
                showSearchContext={showSearchContext}
                versionContext={undefined}
            />
        </div>
    </p>
)

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
        <h3 className={styles.containerTitle}>
            <span className="flex-1">{title}</span>
            {sectionID && (
                <button
                    type="button"
                    className="btn btn-icon"
                    aria-label="Hide Section"
                    onClick={() => onClose?.(sectionID)}
                >
                    <CloseIcon className="icon-inline" />
                </button>
            )}
        </h3>
        <div className={styles.containerContent}>{children}</div>
    </div>
)

interface NoResultsPageProps extends ThemeProps {
    showSearchContext: boolean
}

const videos = [
    {
        title: 'Three ways to search',
        thumbnailPrefix: 'img/watch-and-learn',
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

export const NoResultsPage: React.FunctionComponent<NoResultsPageProps> = ({ showSearchContext, isLightTheme }) => {
    const [hiddenSectionIDs, setHiddenSectionIds] = useTemporarySetting('search.hiddenNoResultsSections')

    const onClose = useCallback(
        sectionID => {
            setHiddenSectionIds((hiddenSectionIDs = []) =>
                !hiddenSectionIDs.includes(sectionID) ? [...hiddenSectionIDs, sectionID] : hiddenSectionIDs
            )
        },
        [setHiddenSectionIds]
    )

    return (
        <>
            <h2>Sourcegraph basics</h2>
            <div className={styles.root}>
                <div className={styles.mainPanels}>
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
                                showSearchContext={showSearchContext}
                                query="repo:sourcegraph const Authentication"
                                patternType={SearchPatternType.regexp}
                            />
                        </Container>
                    )}
                    {!hiddenSectionIDs?.includes(SectionID.COMMON_PROBLEMS) && (
                        <Container sectionID={SectionID.COMMON_PROBLEMS} title="Common Problems" onClose={onClose}>
                            <h4>Finding a specific repository</h4>
                            <p>Repositories are specified by their org/repository-name convention:</p>
                            <SearchInputExample
                                showSearchContext={showSearchContext}
                                query="repo:sourcegraph/about lang:go publish"
                            />
                            <p>
                                To search within all of an org’s repositories, specify only the org name and a trailing
                                slash:
                            </p>
                            <SearchInputExample
                                showSearchContext={showSearchContext}
                                query="repo:sourcegraph/ lang:go publish"
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
                                showSearchContext={showSearchContext}
                                query="repo:sourcegraph/ (lang:typescript OR lang:go) auth"
                            />

                            <h4>Escaping</h4>
                            <p>
                                Because our default mode is literal, escaping requires a dedicated filter. Use the
                                content filter to include spaces and filter keywords in searches.
                            </p>
                            <SearchInputExample
                                showSearchContext={showSearchContext}
                                query={'content:"class Vector"'}
                            />
                        </Container>
                    )}

                    <Container title="More resources">
                        <p>
                            Check out the learn site, including the cheat sheet for more tips on getting the most from
                            Sourcegraph.
                        </p>
                        <p>
                            <Link target="blank" to="https://learn.sourcegraph.com/">
                                Sourcegraph Learn <ExternalLinkIcon className="icon-inline" />
                            </Link>
                            <br />
                            <Link
                                target="blank"
                                to="https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet"
                            >
                                Sourcegraph cheat sheet <ExternalLinkIcon className="icon-inline" />
                            </Link>
                        </p>
                    </Container>

                    {hiddenSectionIDs && hiddenSectionIDs.length > 0 && (
                        <p>
                            Some help panels are hidden.
                            <button type="button" className="btn btn-link" onClick={() => setHiddenSectionIds([])}>
                                Turn all panels on.
                            </button>
                        </p>
                    )}
                </div>
                {!hiddenSectionIDs?.includes(SectionID.VIDEOS) && (
                    <div className={styles.videoPanel}>
                        <Container
                            sectionID={SectionID.VIDEOS}
                            title="Video explanations"
                            className={styles.videoContainer}
                            onClose={onClose}
                        >
                            {videos.map(video => (
                                <ModalVideo
                                    key={video.title}
                                    id={`video-${video.title.toLowerCase().replace(/[^a-z]+/, '-')}`}
                                    title={video.title}
                                    src={video.src}
                                    thumbnail={{
                                        src: `${video.thumbnailPrefix}-${isLightTheme ? 'light' : 'dark'}.png`,
                                        alt: `${video.title} video thumbnail`,
                                    }}
                                    showCaption={true}
                                />
                            ))}
                        </Container>
                    </div>
                )}
            </div>
        </>
    )
}
