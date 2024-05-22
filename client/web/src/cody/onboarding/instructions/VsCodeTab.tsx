import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Link, Icon, Text } from '@sourcegraph/wildcard'

import { EventName } from '../../../util/constants'

import styles from '../CodyOnboarding.module.scss'

export function VSCodeTabInstructions({ telemetryRecorder }: { telemetryRecorder: TelemetryRecorder }): JSX.Element {
    const marketplaceUrl = 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai'

    return (
        <div className={styles.ideContainer}>
            <Text className={classNames('mb-0', styles.install)}>Install</Text>
            <Link
                to="vscode:extension/sourcegraph.cody-ai"
                target="_blank"
                rel="noreferrer"
                className="d-inline-flex align-items-center"
                onClick={event => {
                    event.preventDefault()
                    EVENT_LOGGER.log(EventName.CODY_EDITOR_SETUP_OPEN_MARKETPLACE, {
                        editor: 'VS Code',
                    })
                    telemetryRecorder.recordEvent('cody.onboarding.extension', 'click', {
                        metadata: { vsCode: 1 },
                    })
                    window.location.href = 'vscode:extension/sourcegraph.cody-ai'
                }}
            >
                <div className={styles.ideLinkAlt}> Open in VS Code</div>
                <Icon
                    className={classNames('ml-1', styles.ideLink)}
                    role="img"
                    aria-label="Open in a new tab"
                    svgPath={mdiOpenInNew}
                />
            </Link>
            <div className={styles.ideOr}>or</div>
            <Link
                to={marketplaceUrl}
                target="_blank"
                rel="noreferrer"
                className="d-inline-flex align-items-center"
                onClick={event => {
                    event.preventDefault()
                    EVENT_LOGGER.log(EventName.CODY_EDITOR_SETUP_OPEN_MARKETPLACE, {
                        editor: 'VS Code',
                    })
                    telemetryRecorder.recordEvent('cody.onboarding.openMarketplace', 'click', {
                        metadata: { vsCode: 1 },
                    })
                    window.location.href = marketplaceUrl
                }}
            >
                <div className={classNames(styles.ideLink)}>View in VS Code marketplace</div>
                <Icon
                    className={classNames('ml-1', styles.ideLink)}
                    role="img"
                    aria-label="Open in a new tab"
                    svgPath={mdiOpenInNew}
                />
            </Link>
        </div>
    )
}
