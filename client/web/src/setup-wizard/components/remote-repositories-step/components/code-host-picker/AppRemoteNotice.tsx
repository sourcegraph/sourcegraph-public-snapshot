import type { FC } from 'react'

import { Text } from '@sourcegraph/wildcard'

import styles from './AppRemoteNotice.module.scss'

export const AppRemoteNotice: FC = () => (
    <section>
        <header className={styles.header}>
            <span>Please replace any existing remote code hosts</span>
        </header>
        <Text as="p" className={styles.paragraph}>
            We plan to remove support for remote repositories soon, as remote repositories are cloned locally once
            synced to the app. Please disconnect any remote connections and replace them with local repositories to
            avoid disruption.
        </Text>
    </section>
)
