import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'

import { ComponentCommits } from './ComponentCommits'
import { EntityDetailContentCardProps } from './EntityDetailContent'

interface Props
    extends EntityDetailContentCardProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    entity: CatalogEntityDetailFields
}

export const EntityChangesTab: React.FunctionComponent<Props> = ({
    entity,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
}) => (
    <ComponentCommits
        catalogComponent={entity}
        className="card"
        headerClassName={headerClassName}
        titleClassName={titleClassName}
        bodyClassName={bodyClassName}
        bodyScrollableClassName={bodyScrollableClassName}
    />
)
