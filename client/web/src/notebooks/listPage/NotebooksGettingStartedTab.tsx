import React, { useEffect, useMemo } from 'react'

import classNames from 'classnames'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, Icon, Link } from '@sourcegraph/wildcard'

import { PageRoutes } from '../../routes.constants'

import styles from './NotebooksGettingStartedTab.module.scss'

interface NotebooksGettingStartedTabProps extends TelemetryProps {}

const functionalityPanels = [
    {
        title: 'The command palette',
        description:
            'Use slash commands to choose from the available block Notebook block types. Markdown, file, symbol, and search query blocks are supported.',
        videoSources: [
            {
                type: 'webm',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette.webm',
            },
            {
                type: 'mp4',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette.mp4',
            },
        ],
    },
    {
        title: 'Keep your docs current with symbol blocks',
        description:
            'Symbol blocks follow a chosen symbol anywhere in a file, even as it changes. Create symbol blocks to keep your docs from getting stale.',
        videoSources: [
            {
                type: 'webm',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block.webm',
            },
            {
                type: 'mp4',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block.mp4',
            },
        ],
    },
    {
        title: 'Share Notebooks with your team or company',
        description:
            "Notebooks are private by default, but you can share them with your team (if you're using Sourcegraph organizations) or with your company.",
        videoSources: [
            { type: 'webm', src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing.webm' },
            { type: 'mp4', src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing.mp4' },
        ],
    },
]

export const NotebooksGettingStartedTab: React.FunctionComponent<NotebooksGettingStartedTabProps> = ({
    telemetryService,
}) => {
    useEffect(() => telemetryService.log('NotebooksGettingStartedTabViewed'), [telemetryService])

    const [, setHasSeenGettingStartedTab] = useTemporarySetting('search.notebooks.gettingStartedTabSeen', false)

    useEffect(() => {
        setHasSeenGettingStartedTab(true)
    }, [setHasSeenGettingStartedTab])

    const videoAutoplayAttributes = useMemo(() => {
        const canAutoplay = window.matchMedia('(prefers-reduced-motion: no-preference)').matches
        return canAutoplay ? { autoPlay: true, loop: true, controls: false } : { controls: true }
    }, [])

    return (
        <>
            <Container className="mb-4">
                <div className={classNames(styles.row, 'row')}>
                    <div className="col-12 col-md-6">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            muted={true}
                            playsInline={true}
                            {...videoAutoplayAttributes}
                        >
                            <source
                                type="video/webm"
                                src="https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_overview.webm"
                            />
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_overview.mp4"
                            />
                        </video>
                    </div>
                    <div className="col-12 col-md-6">
                        <h2>Create living documentation effortlessly</h2>
                        <p>
                            Notebooks make creating and sharing knowledge something you'll want to do, not something you
                            avoid.
                        </p>
                        <h3>Use notebooks for&hellip;</h3>
                        <ul className={classNames(styles.narrowList, 'mb-0')}>
                            <li>
                                Onboarding a new teammate: Create{' '}
                                <Link
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    to="https://sourcegraph.com/notebooks/Tm90ZWJvb2s6NDEy"
                                >
                                    focused docs
                                </Link>{' '}
                                that stay up to date
                            </li>
                            <li>Better PR walkthroughs: Quickly link to impacted code that wasn't part of the PR</li>
                            <li>
                                Tracking symbol definitions: Use symbol blocks to ensure you're always reading the
                                latest docs
                            </li>
                            <li>
                                Long-form documentation: ocument complex systems like CI with query blocks and
                                structural search to highlight critical code
                            </li>
                            <li>
                                <Link target="_blank" rel="noopener noreferrer" to="/help/notebooks/notebook-embedding">
                                    Embedding Notebooks
                                </Link>{' '}
                                in existing documentation: view Notebooks content in your existing knowledge management
                                system
                            </li>
                        </ul>
                    </div>
                </div>
            </Container>
            <h3>Example notebooks</h3>
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
            <h3>Using the notepad</h3>
            <Container className="mb-4">
                <div className={classNames(styles.row, 'row', 'mb-4')}>
                    <div className="col-12 col-md-6">
                        <strong>Text</strong>
                    </div>
                    <div className="col-12 col-md-6">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            muted={true}
                            playsInline={true}
                            controls={true}
                        >
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/notebooks/notepad_small_browser.mp4"
                            />
                        </video>
                    </div>
                </div>
            </Container>
            <h3>Functionality</h3>
            <div className={classNames(styles.row, 'row', 'mb-4')}>
                {functionalityPanels.map(panel => (
                    <div key={panel.title} className="col-12 col-md-4">
                        <Container>
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
