import React from 'react'

import { mdiContentSave, mdiContentSaveEditOutline, mdiDelete } from '@mdi/js'

import { pluralize } from '@sourcegraph/common'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './Icons.module.scss'

export const CachedIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tooltip content="A cached result was found for this workspace.">
        <Icon aria-hidden={true} svgPath={mdiContentSave} />
    </Tooltip>
)

export const PartiallyCachedIcon: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({ count }) => {
    const label = `This workspace contains cached results for ${count} ${pluralize('step', count)}.`
    return (
        <Tooltip content={label}>
            <Icon
                role="img"
                className={styles.partiallyCachedIcon}
                aria-hidden={true}
                svgPath={mdiContentSaveEditOutline}
            />
        </Tooltip>
    )
}

export const ExcludeIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tooltip content="Your batch spec was modified to exclude this workspace. Preview again to update.">
        <Icon aria-hidden={true} svgPath={mdiDelete} />
    </Tooltip>
)
