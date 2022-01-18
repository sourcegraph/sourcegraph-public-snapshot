import classNames from 'classnames'
import React, { useCallback } from 'react'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded/src/components/SyntaxHighlightedSearchQuery'
import { ModalVideo } from '@sourcegraph/branded/src/search/documentation/ModalVideo'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { QueryState } from '@sourcegraph/shared/src/search/helpers'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './index.module.scss'
import { SearchExample, exampleQueries } from './SearchExamples'

export interface HomePanelsProps extends TelemetryProps, ThemeProps {
    setQuery: (newState: QueryState) => void
}
interface SearchExamplesProps extends TelemetryProps {
    title: string
    subtitle: string
    examples: SearchExample[]
    icon: JSX.Element
    setQuery: (newState: QueryState) => void
}

const SearchExamples: React.FunctionComponent<SearchExamplesProps> = ({
    title,
    subtitle,
    examples,
    icon,
    telemetryService,
    setQuery,
}) => {
    const searchExampleClicked = useCallback(
        (trackEventName: string, exampleQuery: string) => (): void => {
            setQuery({ query: exampleQuery })
            telemetryService.log(trackEventName)
        },
        [setQuery, telemetryService]
    )
    return (
        <div className={styles.searchExamplesWrapper}>
            <div className={classNames('d-flex align-items-baseline mb-2', styles.searchExamplesTitleWrapper)}>
                <div className={classNames('mr-2', styles.title, styles.searchExamplesTitle)}>{title}</div>
                <div className="font-weight-normal text-muted">{subtitle}</div>
            </div>
            <div className={styles.searchExamples}>
                {examples.map(example => (
                    <div key={example.query} className="search-example-card-wrapper">
                        <Link
                            to={example.to}
                            className={classNames('card', styles.searchExampleCard)}
                            onClick={searchExampleClicked(example.trackEventName, example.query)}
                        >
                            <div className={classNames('search-example-example-icons', styles.searchExampleIcon)}>
                                {icon}
                            </div>
                            <div className={styles.searchExampleQueryWrapper}>
                                <div className={styles.searchExampleQuery}>
                                    <SyntaxHighlightedSearchQuery query={example.query} />
                                </div>
                            </div>
                        </Link>
                        <p>{example.label}</p>
                    </div>
                ))}
            </div>
        </div>
    )
}

export const HomePanels: React.FunctionComponent<HomePanelsProps> = props => (
    <>
        <div className={styles.vsceSearchHomepage}>
            <div className={styles.helpContent}>
                <SearchExamples
                    title="Search examples"
                    subtitle="Find answers faster with code search across multiple repos and commits"
                    examples={exampleQueries}
                    icon={<MagnifyingGlassSearchIcon />}
                    {...props}
                />
                <div className={styles.thumbnail}>
                    <div className={classNames(styles.title, 'mb-2')}>Watch and learn</div>
                    <ModalVideo
                        id="three-ways-to-search-title"
                        title="Three ways to search"
                        src="https://www.youtube-nocookie.com/embed/XLfE2YuRwvw"
                        thumbnail={{
                            src: `img/watch-and-learn-${props.isLightTheme ? 'light' : 'dark'}.png`,
                            alt: 'Watch and learn video thumbnail',
                        }}
                        onToggle={isOpen =>
                            props.telemetryService.log(
                                isOpen ? 'HomepageVideoWaysToSearchClicked' : 'HomepageVideoClosed'
                            )
                        }
                        assetsRoot="https://sourcegraph.com/.assets/"
                    />
                </div>
            </div>

            <div className="d-flex justify-content-center">
                <div className={classNames('card', styles.ctaCard)}>
                    <div className="d-flex align-items-center">
                        <span className="badge badge-merged text-uppercase mr-2">Beta</span>
                    </div>
                    <span>
                        Search your public and private code. Read our{' '}
                        <a
                            href="https://about.sourcegraph.com/blog/why-index-the-oss-universe/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            blog post
                        </a>{' '}
                        to learn more.
                    </span>
                </div>
            </div>
        </div>
    </>
)

const MagnifyingGlassSearchIcon = React.memo(() => (
    <svg width="18" height="18" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            d="M6.686.5a6.672 6.672 0 016.685 6.686 6.438 6.438 0 01-1.645 4.32l.308.308h.823L18 16.957 16.457 18.5l-5.143-5.143v-.823l-.308-.308a6.438 6.438 0 01-4.32 1.645A6.672 6.672 0 010 7.186 6.672 6.672 0 016.686.5zm0 2.057a4.61 4.61 0 00-4.629 4.629 4.61 4.61 0 004.629 4.628 4.61 4.61 0 004.628-4.628 4.61 4.61 0 00-4.628-4.629z"
            fill="currentColor"
        />
    </svg>
))
