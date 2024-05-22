import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Icon, Link, Text } from '@sourcegraph/wildcard'

import { EventName } from '../../../util/constants'

import styles from '../CodyOnboarding.module.scss'

export function NeoVimTabInstructions({ telemetryRecorder }: { telemetryRecorder: TelemetryRecorder }): JSX.Element {
    const marketplaceUrl = 'https://github.com/sourcegraph/sg.nvim#setup'

    return (
        <div className={styles.ideContainer}>
            <Text className={classNames('mb-0', styles.install)}>Install</Text>
            <Link
                to={marketplaceUrl}
                target="_blank"
                rel="noreferrer"
                className="d-inline-flex align-items-center"
                onClick={event => {
                    event.preventDefault()
                    EVENT_LOGGER.log(EventName.CODY_EDITOR_SETUP_OPEN_MARKETPLACE, {
                        editor: 'NeoVim',
                    })
                    telemetryRecorder.recordEvent('cody.onboarding.openMarketplace', 'click', {
                        metadata: { neoVim: 1 },
                    })
                    window.location.href = marketplaceUrl
                }}
            >
                <div className={styles.ideLink}>View the GitHub repo</div>
                <Icon
                    className={classNames('ml-1', styles.ideLink)}
                    role="img"
                    aria-label="Open in a new tab"
                    svgPath={mdiOpenInNew}
                />
            </Link>
            <Text className={classNames('mb-0', styles.ideTextLight)}>
                Follow the instructions in <br /> the readme.md
            </Text>
        </div>
    )
}
