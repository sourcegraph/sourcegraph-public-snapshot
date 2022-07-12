import React from 'react'

import { mdiContentSave, mdiContentSaveEditOutline, mdiDelete } from '@mdi/js'

import { pluralize } from '@sourcegraph/common'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './Icons.module.scss'

export const CachedIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tooltip content="A cached result was found for this workspace.">
        <Icon aria-label="A cached result was found for this workspace." svgPath={mdiContentSave} />
    </Tooltip>
)

export const PartiallyCachedIcon: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({ count }) => {
    const label = `This workspace contains cached results for ${count} ${pluralize('step', count)}.`
    return (
        <Tooltip content={label}>
            <Icon
                aria-label={label}
                role="img"
                className={styles.partiallyCachedIcon}
                svgPath={mdiContentSaveEditOutline}
            />
        </Tooltip>
    )
}

export const ExcludeIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tooltip content="Your batch spec was modified to exclude this workspace. Preview again to update.">
        <Icon
            aria-label="Your batch spec was modified to exclude this workspace. Preview again to update."
            svgPath={mdiDelete}
        />
    </Tooltip>
)
