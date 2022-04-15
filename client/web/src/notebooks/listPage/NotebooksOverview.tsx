import React, { useMemo } from 'react'

import classNames from 'classnames'

import { Container } from '@sourcegraph/wildcard'

import styles from './NotebooksOverview.module.scss'

export const NotebooksOverview: React.FunctionComponent = () => {
    const videoAutoplayAttributes = useMemo(() => {
        const canAutoplay = window.matchMedia('(prefers-reduced-motion: no-preference)').matches
        return canAutoplay ? { autoPlay: true, loop: true, controls: false } : { controls: true }
    }, [])

    return (
        <Container className="mb-4">
            <div className={classNames(styles.row, 'row')}>
                <div className="col-12 col-md-7">
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
