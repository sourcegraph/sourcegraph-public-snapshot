import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Link, Icon, Text } from '@sourcegraph/wildcard'

import { EventName } from '../../../util/constants'

import styles from '../CodyOnboarding.module.scss'

export function JetBrainsTabInstructions({
    telemetryRecorder,
}: {
    onBack?: () => void

    telemetryRecorder: TelemetryRecorder
}): JSX.Element {
    const marketplaceUrl = 'https://plugins.jetbrains.com/plugin/9682-sourcegraph-cody--code-search'

    const icons = [
        { id: 1, icon: 'PyCharm' },
        { id: 2, icon: 'PhpStorm' },
        { id: 3, icon: 'WebStorm' },
        { id: 4, icon: 'RubyMine' },
        { id: 5, icon: 'GoLand' },
        { id: 6, icon: 'RubyMine' },
    ]

    return (
        <div className="w-100">
            <div className={styles.ideContainer}>
                <Text className={classNames('mb-0', styles.install)}>Install</Text>
                <div className={classNames('d-flex flex-column gap-0', styles.ideTextLight)}>
                    <Text className="mb-0">1. Open your IDE</Text>
                    <Text className="mb-0">2. {'Settings > plugins (CMD + ,)'}</Text>
                    <Text className="mb-0">3. Search “Sourcegraph”</Text>
                </div>
                <div className={styles.ideOr}>or</div>
                <Link
                    to={marketplaceUrl}
                    target="_blank"
                    rel="noopener"
                    className="d-inline-flex align-items-center"
                    onClick={event => {
                        event.preventDefault()
                        EVENT_LOGGER.log(EventName.CODY_EDITOR_SETUP_OPEN_MARKETPLACE, {
                            editor: 'JetBrains',
                        })
                        telemetryRecorder.recordEvent('cody.onboarding.openMarketplace', 'click', {
                            metadata: { jetBrains: 1 },
                        })
                        window.location.href = marketplaceUrl
                    }}
                >
                    <div className={styles.ideLink}> Install from the marketplace</div>
                    <Icon
                        className={classNames('ml-1', styles.ideLink)}
                        role="img"
                        aria-label="Open in a new tab"
                        svgPath={mdiOpenInNew}
                    />
                </Link>
            </div>
            <div className={classNames('w-100', styles.ideWorks)}>
                <Text>Works with:</Text>
                <div className={styles.ideLogoContainer}>
                    {icons.map(icon => (
                        <img
                            key={icon.id}
                            src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${icon.icon}.svg`}
                            alt="PS"
                            aria-hidden={true}
                            width={36.725}
                            height={36.725}
                        />
                    ))}
                </div>
            </div>
        </div>
    )
}
