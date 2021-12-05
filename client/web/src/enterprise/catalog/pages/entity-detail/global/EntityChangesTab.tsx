import classNames from 'classnames'
import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'

import { ComponentCommits } from './EntityCommits'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: CatalogEntityDetailFields
    className?: string
}

export const EntityChangesTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className={classNames('container my-3', className)}>
        <div className="card">
            <ComponentCommits catalogComponent={entity} />
        </div>
    </div>
)
