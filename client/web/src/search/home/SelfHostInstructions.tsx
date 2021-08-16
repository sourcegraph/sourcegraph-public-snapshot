import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useState } from 'react'

import styles from './SelfHostInstructions.module.scss'

export const SelfHostInstructions: React.FunctionComponent<{}> = () => {
    const dockerCommand =
        'docker run --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.30.3'

    const copyTooltip = 'Copy command'
    const copyCompletedTooltip = 'Copied!'

    const [currentCopyTooltip, setCurrentCopyTooltip] = useState(copyTooltip)

    const onCopy = (): void => {
        copy(dockerCommand)
        setCurrentCopyTooltip(copyCompletedTooltip)
        setTimeout(() => setCurrentCopyTooltip(copyTooltip), 1000)
    }

    return (
        <div className={styles.wrapper}>
            <div className={styles.column}>
                <h2>
                    <DownloadIcon className={classNames('icon-inline mr-2', styles.downloadIcon)} /> Self-hosted
                    deployment
                </h2>
                <ul className={styles.featureList}>
                    <li>Free for up to 10 users</li>
                    <li>Supports additional (and local) code hosts</li>
                    <li>Team oriented functionality</li>
                    <li>Your code never leaves your server</li>
                    <li>Free 30 day trial of enterprise-only features</li>
                </ul>
                <a href="https://docs.sourcegraph.com/self-hosted-vs-cloud" target="_blank" rel="noopener noreferrer">
                    Learn more about self-hosted vs. cloud features <OpenInNewIcon className="icon-inline" />{' '}
                    <span className="sr-only">(Open in new window)</span>
                </a>
            </div>

            <div className={styles.column}>
                <div>
                    <strong>Quickstart:</strong> launch Sourcegraph at{' '}
                    <a href="http://localhost:3370" target="_blank" rel="noopener noreferrer">
                        http://localhost:3370
                    </a>
                </div>
                <div className={styles.codeWrapper}>
                    <button
                        type="button"
                        className={classNames('btn btn-link', styles.copyButton)}
                        onClick={onCopy}
                        data-tooltip={currentCopyTooltip}
                        data-placement="top"
                    >
                        <ContentCopyIcon className="icon-inline" />
                        <span className="sr-only">Copy to clipboard</span>
                    </button>
                    <code className={styles.code}>{dockerCommand}</code>
                </div>
                <div className="d-flex justify-content-between">
                    <a href="https://docs.sourcegraph.com/admin/install" target="_blank" rel="noopener noreferrer">
                        Learn how to deploy a server or cluster <OpenInNewIcon className="icon-inline" />{' '}
                        <span className="sr-only">(Open in new window)</span>
                    </a>
                    <a href="https://about.sourcegraph.com/contact/request-info/">Talk to an engineer</a>
                </div>
            </div>
        </div>
    )
}
