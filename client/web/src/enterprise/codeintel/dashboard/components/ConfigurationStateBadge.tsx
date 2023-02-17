import { FunctionComponent } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

import styles from './ConfigurationStateBadge.module.scss'

export interface IndexerDescription {
    key: string
    name: string
    url: string
}

export interface ConfigurationStateBadgeProps {
    indexer: IndexerDescription
    className?: string
}

export const ConfigurationStateBadge: FunctionComponent<ConfigurationStateBadgeProps> = ({ indexer, className }) => (
    <small className={classNames(className, styles.hint)}>
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-muted" />{' '}
        <strong>
            Configure {indexer.key} via {indexer.name}?
        </strong>
    </small>
)
