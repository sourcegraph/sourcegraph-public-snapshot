import { FC } from 'react'

import styles from './RemoteNotice.module.scss'

export const RemoteNotice: FC = () => (
    <section>
        <header className={styles.header}>
            <span>Please replace any existing remote code hosts</span>
        </header>
        <p className={styles.paragraph}>
            We plan to remove support for remote repositories soon, as remote repositories are cloned locally once
            synced to the app. Please disconnect any remote connections and replace them with local repositories to
            avoid disruption.
        </p>
    </section>
)
