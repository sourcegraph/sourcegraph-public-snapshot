import React, { useState } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container } from '@sourcegraph/wildcard'

import styles from './NotebooksGettingStarted.module.scss'

interface NotebooksGettingStartedProps extends TelemetryProps {}

const panels = [
    {
        title: 'Enable the notepad',
        description:
            "Create notebooks while you browse. Add searches, files, and file ranges without leaving the page you're on, then create a notebook with one click.",
        videoSources: [
            { type: 'mp4', src: 'https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.mp4' },
        ],
    },
    {
        title: 'The command palette',
        description:
            'Use slash commands to choose from the available block Notebook block types. Markdown, file, symbol, and search query blocks are supported.',
        videoSources: [
            {
                type: 'mp4',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette.mp4',
            },
        ],
    },
    {
        title: 'Keep your docs current-automatically-with symbol blocks',
        description:
            'Symbol blocks follow a chosen symbol anywhere in a file, even as it changes. Create symbol blocks to keep your docs from getting stale.',
        videoSources: [
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
            { type: 'mp4', src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing.mp4' },
        ],
    },
]

export const NotebooksGettingStarted: React.FunctionComponent<NotebooksGettingStartedProps> = ({
    telemetryService,
}) => {
    const [selectedPanelIndex, setSelectedPanelIndex] = useState(0)
    const selectedPanel = panels[selectedPanelIndex]

    const previousPanelIndex = (): number => (selectedPanelIndex === 0 ? panels.length - 1 : selectedPanelIndex - 1)
    const nextPanelIndex = (): number => (selectedPanelIndex === panels.length - 1 ? 0 : selectedPanelIndex + 1)

    const selectPanelIndex = (index: number): void => {
        const title = panels[index].title
        telemetryService.log('SearchNotebooksGettingStartedPanelViewed', { title }, { title })
        setSelectedPanelIndex(index)
    }

    return (
        <Container className="mb-4">
            <div className={classNames(styles.row, 'row')}>
                <div className="col-12 col-md-7">
                    {/* To optimize switching between panels we pre-render all of the videos and switch between them using d-block/d-none.
                    Otherwise we get flashing content when one video is unmounted and another video is mounted. */}
                    {panels.map((panel, panelIndex) => (
                        <video
                            key={`panel-${panel.title}`}
                            className={classNames(
                                'w-100',
                                'h-auto',
                                'shadow',
                                'percy-hide',
                                selectedPanelIndex === panelIndex ? 'd-block' : 'd-none'
                            )}
                            width={1280}
                            height={720}
                            autoPlay={true}
                            muted={true}
                            loop={true}
                            playsInline={true}
                            controls={false}
                        >
                            {panel.videoSources.map(videoSource => (
                                <source
                                    key={videoSource.src}
                                    type={`video/${videoSource.type}`}
                                    src={videoSource.src}
                                />
                            ))}
                        </video>
                    ))}
                </div>
                <div className={classNames('col-12', 'col-md-5', styles.right)}>
                    <div className="mt-2">
                        <h2>{selectedPanel.title}</h2>
                        <p className={styles.description}>{selectedPanel.description}</p>
                    </div>
                    <div className={styles.panelButtons}>
                        <div>
                            {range(panels.length).map(panelIndex => (
                                <Button
                                    className={classNames(
                                        styles.panelButton,
                                        selectedPanelIndex === panelIndex && styles.selectedPanelButton
                                    )}
                                    key={`getting_started_panel_${panelIndex}`}
                                    onClick={() => selectPanelIndex(panelIndex)}
                                >
                                    {panelIndex + 1}
                                </Button>
                            ))}
                        </div>
                        <div>
                            <Button
                                className="mr-2"
                                variant="secondary"
                                outline={true}
                                onClick={() => selectPanelIndex(previousPanelIndex())}
                            >
                                Previous
                            </Button>
                            <Button variant="primary" onClick={() => selectPanelIndex(nextPanelIndex())}>
                                Next
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        </Container>
    )
}
