import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import * as React from 'react'

import { Icon } from '@sourcegraph/wildcard'

/** Displays the extension's configuration state (not added, added and enabled, added and disabled). */
export const ExtensionConfigurationState: React.FunctionComponent<{
    isAdded: boolean
    isEnabled: boolean

    enabledIconOnly?: boolean

    className?: string
}> = ({ isAdded, isEnabled, enabledIconOnly, className = '' }) =>
    isAdded && isEnabled ? (
        <span className={classNames('text-success', className)}>
            <Icon as={CheckIcon} /> {!enabledIconOnly && 'Enabled'}
        </span>
    ) : (
        <span className={classNames('text-muted', className)}>Disabled</span>
    )
