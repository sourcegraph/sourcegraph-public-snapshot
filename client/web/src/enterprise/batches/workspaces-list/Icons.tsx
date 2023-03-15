import React from 'react'

import { mdiContentSave, mdiContentSaveEditOutline, mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './Icons.module.scss'

export const CachedIcon: React.FunctionComponent<React.PropsWithChildren<{ cacheDisabled?: boolean }>> = ({
    cacheDisabled,
}) => {
    const label = `A cached result was found for this workspace${cacheDisabled ? ', but will not be used' : ''}.`
    return (
        <Tooltip content={label}>
            <Icon
                tabIndex={0}
                role="tooltip"
                aria-label={label}
                className={classNames(cacheDisabled && 'text-muted')}
                svgPath={mdiContentSave}
            />
        </Tooltip>
    )
}

export const PartiallyCachedIcon: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; cacheDisabled?: boolean }>
> = ({ count, cacheDisabled }) => {
    const label = `This workspace contains cached results for ${count} ${pluralize('step', count)}${
        cacheDisabled ? ', but will not be used' : ''
    }.`
    return (
        <Tooltip content={label}>
            <Icon
                aria-label={label}
                tabIndex={0}
                role="tooltip"
                className={classNames(styles.partiallyCachedIcon, cacheDisabled && 'text-muted')}
                svgPath={mdiContentSaveEditOutline}
            />
        </Tooltip>
    )
}

export const ExcludeIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tooltip content="Your batch spec was modified to exclude this workspace. Preview again to update.">
        <Icon
            aria-label="Your batch spec was modified to exclude this workspace. Preview again to update."
            tabIndex={0}
            role="tooltip"
            svgPath={mdiDelete}
        />
    </Tooltip>
)
