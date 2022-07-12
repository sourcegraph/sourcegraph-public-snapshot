import * as React from 'react'

import { mdiCheck } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

/** Displays the extension's configuration state (not added, added and enabled, added and disabled). */
export const ExtensionConfigurationState: React.FunctionComponent<
    React.PropsWithChildren<{
        isAdded: boolean
        isEnabled: boolean

        enabledIconOnly?: boolean

        className?: string
    }>
> = ({ isAdded, isEnabled, enabledIconOnly, className = '' }) =>
    isAdded && isEnabled ? (
        <span className={classNames('text-success', className)}>
            <Icon aria-hidden={true} svgPath={mdiCheck} /> {!enabledIconOnly && 'Enabled'}
        </span>
    ) : (
        <span className={classNames('text-muted', className)}>Disabled</span>
    )
