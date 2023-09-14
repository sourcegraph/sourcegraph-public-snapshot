import type { FC } from 'react'

import { mdiGit } from '@mdi/js'

import { Text, Icon } from '@sourcegraph/wildcard'

import styles from './AppZeroStates.module.scss'

export const NoReposAddedState: FC = () => (
    <div className={styles.noRepoState}>
        <Icon svgPath={mdiGit} aria-hidden={true} className={styles.noRepoStateIcon} />

        <Text className="mb-0">Import your first repository</Text>
    </div>
)
