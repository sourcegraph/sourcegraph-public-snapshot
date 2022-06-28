import React from 'react'

import ContentSaveEditOutlineIcon from 'mdi-react/ContentSaveEditOutlineIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'

import { pluralize } from '@sourcegraph/common'
import { Icon } from '@sourcegraph/wildcard'

import styles from './Icons.module.scss'

export const CachedIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Icon
        data-tooltip="A cached result was found for this workspace."
        aria-label="A cached result was found for this workspace."
        as={ContentSaveIcon}
    />
)

export const PartiallyCachedIcon: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({ count }) => {
    const label = `This workspace contains cached results for ${count} ${pluralize('step', count)}.`
    return (
        <Icon
            role="img"
            className={styles.partiallyCachedIcon}
            data-tooltip={label}
            aria-label={label}
            as={ContentSaveEditOutlineIcon}
        />
    )
}

export const ExcludeIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Icon
        data-tooltip="Your batch spec was modified to exclude this workspace. Preview again to update."
        aria-label="Your batch spec was modified to exclude this workspace. Preview again to update."
        as={DeleteIcon}
    />
)
