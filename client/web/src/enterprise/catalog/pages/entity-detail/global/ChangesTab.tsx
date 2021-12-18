import classNames from 'classnames'
import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ComponentStateDetailFields } from '../../../../../graphql-operations'

import { ComponentCommits } from './ComponentCommits'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: ComponentStateDetailFields
    className?: string
}

export const ChangesTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className={classNames('container my-3', className)}>
        <div className="card">
            <ComponentCommits component={entity} />
        </div>
    </div>
)
