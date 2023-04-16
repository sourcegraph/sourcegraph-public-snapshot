import React from 'react'

import { mdiShimmer } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '../../utils/Icon'

import styles from './Followups.module.css'

export const Followups: React.FunctionComponent<{
    followups: string[]
    onFollowupSelect: (followup: string) => void
    className?: string
}> = ({ followups, onFollowupSelect, className }) =>
    followups.length > 0 ? (
        <ul className={classNames(className, styles.container)}>
            {followups.map((followup, index) => (
                // eslint-disable-next-line react/no-array-index-key
                <li key={index}>
                    <button type="button" className={styles.item} onClick={() => onFollowupSelect(followup)}>
                        <Icon svgPath={mdiShimmer} className={styles.icon} /> {followup}
                    </button>
                </li>
            ))}
        </ul>
    ) : null
