import { FC } from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { H3, Link, Text, Icon } from '@sourcegraph/wildcard'

import { ShieldLogo } from './ShieldImg'

import styles from './SentinelBanner.module.scss'

export const SentinelBanner: FC = () => (
    <div className={styles.container}>
        <div>
            <ShieldLogo />
        </div>
        <div>
            <H3 className={styles.title}>
                Sourcegraph now identifies <strong>code vulnerabilities</strong>
            </H3>
            <Text className={styles.text}>
                This is <strong>Sourcegraph Sentinel</strong> running on a few samples repos. Sentinel is coming soon on
                Sourcegraph Enterprise.
            </Text>
            <div>
                <Link
                    className={styles.learnMore}
                    to="/help/code_insights/references/license"
                    rel="noopener noreferrer"
                    target="_blank"
                >
                    Learn more
                </Link>
                <Link
                    className={styles.text}
                    to="/help/code_insights/references/license"
                    rel="noopener noreferrer"
                    target="_blank"
                >
                    Community Discussion <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                </Link>
            </div>
        </div>
    </div>
)
