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
            'Create Notebooks faster with the notepad. Add searches, files, and file ranges to the notepad and create a notebook with one click.',
    },
    {
        title: 'Enable the notepad 2',
        description:
            'Create Notebooks faster with the notepad. Add searches, files, and file ranges to the notepad and create a notebook with one click.',
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
                    <div>
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            width={1280}
                            height={720}
                            autoPlay={true}
                            muted={true}
                            loop={true}
                            playsInline={true}
                            controls={false}
                        >
                            <source
                                type="video/webm"
                                src="https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.webm"
                            />
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.mp4"
                            />
                        </video>
                    </div>
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
                <div className="col-12 col-md-5">
                    <h2>Automate large-scale code changes</h2>
                    <p>
                        Batch Changes gives you a declarative structure for finding and modifying code across all of
                        your repositories. Its simple UI makes it easy to track and manage all of your changesets
                        through checks and code reviews until each change is merged.
                    </p>
                    <h3>Common use cases</h3>
                    <ul className={classNames(styles.narrowList, 'mb-0')}>
                        <li>Update configuration files across many repositories</li>
                        <li>Update libraries which call your APIs</li>
                        <li>Rapidly fix critical security issues</li>
                        <li>Update boilerplate code</li>
                        <li>Pay down tech debt</li>
                    </ul>
                </div>
            </div>
        </Container>
    )
}
