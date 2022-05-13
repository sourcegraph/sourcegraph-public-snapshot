import React, { useState } from 'react'

import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'

import styles from './SelfHostInstructions.module.scss'

export const SelfHostInstructions: React.FunctionComponent<React.PropsWithChildren<TelemetryProps>> = ({
    telemetryService,
}) => {
    const dockerCommand =
        'docker run --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.39.1'

    const copyTooltip = 'Copy command'
    const copyCompletedTooltip = 'Copied!'

    const [currentCopyTooltip, setCurrentCopyTooltip] = useState(copyTooltip)

    const onCopy = (): void => {
        telemetryService.log('HomepageCTAClicked', { campaign: 'Local install' }, { campaign: 'Local install' })
        copy(dockerCommand)
        setCurrentCopyTooltip(copyCompletedTooltip)
        setTimeout(() => setCurrentCopyTooltip(copyTooltip), 1000)
    }

    const onTalkToEngineerClicked = (): void => {
        telemetryService.log(
            'HomepageCTAClicked',
            { campaign: 'Talk to an engineer' },
            { campaign: 'Talk to an engineer' }
        )
    }

    return (
        <div className={styles.wrapper}>
            <div className={styles.column}>
                <Typography.H2>
                    <Icon className={classNames('mr-2', styles.downloadIcon)} as={DownloadIcon} /> Self-hosted
                    deployment
                </Typography.H2>
                <ul className={styles.featureList}>
                    <li>Free for up to 10 users</li>
                    <li>Supports additional (and local) code hosts</li>
                    <li>Team oriented functionality</li>
                    <li>Your code never leaves your server</li>
                    <li>Free 30 day trial of enterprise-only features</li>
                </ul>
                <Link
                    to="https://docs.sourcegraph.com/cloud/cloud_ent_on-prem_comparison"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Learn more about self-hosted vs. cloud features{' '}
                    <Icon aria-label="Open in new window" as={OpenInNewIcon} />
                </Link>
            </div>

            <div className={styles.column}>
                <div>
                    <strong>Quickstart:</strong> launch Sourcegraph at http://localhost:7080
                </div>
                <MarketingBlock wrapperClassName={styles.codeWrapper} contentClassName={styles.codeContent}>
                    <Button
                        className={styles.copyButton}
                        onClick={onCopy}
                        data-tooltip={currentCopyTooltip}
                        data-placement="top"
                        aria-label="Copy Docker command to clipboard"
                        variant="link"
                    >
                        <Icon as={ContentCopyIcon} />
                    </Button>
                    <code className={styles.codeBlock}>{dockerCommand}</code>
                </MarketingBlock>
                <div className="d-flex justify-content-between">
                    <Link
                        to="https://docs.sourcegraph.com/admin/install"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="mr-2"
                    >
                        Learn how to deploy a server or cluster{' '}
                        <Icon aria-label="Open in new window" as={OpenInNewIcon} />
                    </Link>
                    <Link
                        to="https://info.sourcegraph.com/talk-to-a-developer?form_submission_source=inproduct&utm_campaign=inproduct-self-hosted-install&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=self-hosted-install"
                        onClick={onTalkToEngineerClicked}
                        className="text-right flex-shrink-0"
                    >
                        Talk to an engineer
                    </Link>
                </div>
            </div>
        </div>
    )
}
