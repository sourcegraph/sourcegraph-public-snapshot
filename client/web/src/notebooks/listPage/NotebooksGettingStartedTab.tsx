import React, { useEffect, useMemo } from 'react'

import classNames from 'classnames'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, Icon, Link, Typography } from '@sourcegraph/wildcard'

import { PageRoutes } from '../../routes.constants'
import { ThemePreference } from '../../stores/themeState'
import { useTheme } from '../../theme'

import styles from './NotebooksGettingStartedTab.module.scss'

interface NotebooksGettingStartedTabProps extends TelemetryProps {}

const functionalityPanels = [
    {
        title: 'Keep your docs current with symbol blocks',
        description:
            'Symbol blocks follow a chosen symbol anywhere in a file, even as it changes. Create symbol blocks to keep your docs from getting stale.',
        image: {
            light: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block_light.png',
            dark: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block_dark.png',
            alt: 'Notebook symbol block',
        },
    },
    {
        title: 'The command palette',
        description:
            'Use slash commands to choose from the available block Notebook block types. Markdown, file, symbol, and search query blocks are supported.',
        image: {
            light: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette_light.png',
            dark: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette_dark.png',
            alt: 'Notebooks command pallete',
        },
    },
    {
        title: 'Share Notebooks with your team or company',
        description:
            "Notebooks are private by default, but you can share them with your team (if you're using Sourcegraph organizations) or with your company.",
        image: {
            light: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing_light.png',
            dark: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing_dark.png',
            alt: 'Notebooks sharing dialog',
        },
    },
]

export const NotebooksGettingStartedTab: React.FunctionComponent<
    React.PropsWithChildren<NotebooksGettingStartedTabProps>
> = ({ telemetryService }) => {
    useEffect(() => telemetryService.log('NotebooksGettingStartedTabViewed'), [telemetryService])

    const [, setHasSeenGettingStartedTab] = useTemporarySetting('search.notebooks.gettingStartedTabSeen', false)

    useEffect(() => {
        setHasSeenGettingStartedTab(true)
    }, [setHasSeenGettingStartedTab])

    const videoAutoplayAttributes = useMemo(() => {
        const canAutoplay = window.matchMedia('(prefers-reduced-motion: no-preference)').matches
        return canAutoplay ? { autoPlay: true, loop: true, controls: false } : { controls: true }
    }, [])

    const isLightTheme = useTheme().enhancedThemePreference === ThemePreference.Light

    return (
        <>
            <Container className="mb-4">
                <div className={classNames(styles.row, 'row')}>
                    <div className="col-12 col-md-6">
                        <video
                            key={`notebooks_overview_video_${isLightTheme}`}
                            className="w-100 h-auto shadow percy-hide"
                            muted={true}
                            playsInline={true}
                            {...videoAutoplayAttributes}
                        >
                            <source
                                type="video/webm"
                                src={`https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_overview_v3_${
                                    isLightTheme ? 'light' : 'dark'
                                }.webm`}
                            />
                            <source
                                type="video/mp4"
                                src={`https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_overview_v3_${
                                    isLightTheme ? 'light' : 'dark'
                                }.mp4`}
                            />
                        </video>
                    </div>
                    <div className="col-12 col-md-6">
                        <Typography.H2>Create living documentation effortlessly</Typography.H2>
                        <p>
                            Notebooks make creating and sharing knowledge something you'll want to do, not something you
                            avoid.
                        </p>
                        <Typography.H3>Use notebooks to&hellip;</Typography.H3>
                        <ul className={classNames(styles.narrowList, 'mb-0')}>
                            <li className="mb-1">Create focused onboarding docs that stay up to date</li>
                            <li className="mb-1">Prepare pull request walkthroughs for your teammates</li>
                            <li className="mb-1">
                                Document complex systems to make them more approachable to new engineers
                            </li>
                            <li className="mb-1">
                                Track symbol definitions to ensure you're always reading the latest docs
                            </li>
                            <li className="mb-1">
                                <Link target="_blank" rel="noopener noreferrer" to="/help/notebooks/notebook-embedding">
                                    Embed
                                </Link>{' '}
                                the most current code anywhere you host your docs
                            </li>
                        </ul>
                    </div>
                </div>
            </Container>
            <Typography.H3>Example notebooks</Typography.H3>
            <div className={classNames(styles.row, 'row', 'mb-4')}>
                <div className="col-12 col-md-6">
                    <Container>
                        <Link
                            target="_blank"
                            rel="noopener noreferrer"
                            to="https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MQ=="
                        >
                            Find Log4J dependencies <Icon as={OpenInNewIcon} />
                        </Link>
                        <div className="mt-2">Find Log4J dependencies across all your code.</div>
                    </Container>
                </div>
                <div className="col-12 col-md-6">
                    <Container>
                        <Link
                            target="_blank"
                            rel="noopener noreferrer"
                            to="https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTM="
                        >
                            Learn Sourcegraph / Find code across all of your repositories <Icon as={OpenInNewIcon} />
                        </Link>
                        <div className="mt-2">Learn how to find and reference code across all your repositories.</div>
                    </Container>
                </div>
            </div>
            <Typography.H3>Powerful creation features</Typography.H3>
            <Container className="mb-4">
                <div className={classNames(styles.row, 'row', 'mb-4')}>
                    <div className="col-12 col-md-6">
                        <strong>Enable the notepad for frictionless knowledge sharing</strong>
                        <p className="mt-1">
                            With the notepad, create notebooks while you browse. Add searches, files, and file ranges
                            without leaving the page you're on, then create a notebook of it all with one click.
                        </p>
                        <strong>Compose rich documentation with multiple block types</strong>
                        <p className="mt-1">
                            Create text content with Markdown blocks, track symbols within files with symbol blocks, and
                            add whole files or line ranges with file blocks.
                        </p>
                    </div>
                    <div className="col-12 col-md-6">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            muted={true}
                            playsInline={true}
                            controls={true}
                            src={`https://storage.googleapis.com/sourcegraph-assets/notebooks/notepad_overview_${
                                isLightTheme ? 'light' : 'dark'
                            }.mp4`}
                        />
                    </div>
                </div>
            </Container>
            <Typography.H3>Functionality</Typography.H3>
            <div className={classNames(styles.row, 'row', 'mb-4')}>
                {functionalityPanels.map(panel => (
                    <div key={panel.title} className="col-12 col-md-4">
                        <Container>
                            <img
                                className="w-100"
                                src={isLightTheme ? panel.image.light : panel.image.dark}
                                alt={panel.image.alt}
                            />
                            <div className="my-2">
                                <strong>{panel.title}</strong>
                            </div>
                            <p>{panel.description}</p>
                        </Container>
                    </div>
                ))}
            </div>
            <div className={classNames(styles.row, 'row', 'mb-4')}>
                <div className="col-12 col-md-6">
                    <div className="mb-2">
                        <strong>Ready to get started?</strong>
                    </div>
                    <div className="mb-2">
                        Notebooks can be used for onboarding, documentation, incident response, and more.
                    </div>
                    <Link to={PageRoutes.NotebookCreate}>Create a notebook</Link>
                </div>
                <div className="col-12 col-md-6">
                    <div className="mb-2">
                        <strong>Learn more about notebooks</strong>
                    </div>
                    <div className="mb-2">Read in-depth material about all of notebooks' features.</div>
                    <Link target="_blank" rel="noopener noreferrer" to="/help/notebooks">
                        Documentation <Icon as={OpenInNewIcon} />
                    </Link>
                </div>
            </div>
        </>
    )
}
