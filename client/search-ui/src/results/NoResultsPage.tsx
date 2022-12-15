import React, { useCallback, useEffect } from 'react'

import { mdiClose, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { SearchContextProps } from '@sourcegraph/search'
import { NoResultsSectionID as SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Link, Icon, H2, H3, Text } from '@sourcegraph/wildcard'

import { ModalVideo } from '../documentation/ModalVideo'

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
