import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { ComponentSources } from './ComponentSources'
import { EntityDetailContentCardProps } from './EntityDetailContent'
import { EntityOwners } from './EntityOwners'

interface Props
    extends EntityDetailContentCardProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    entity: CatalogEntityDetailFields
}

export const EntityCodeTab: React.FunctionComponent<Props> = ({
    entity,
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
    ...props
}) => (
    <div>
        {entity.__typename === 'CatalogComponent' && (
            <ComponentSourceDefinitions catalogComponent={entity} className="mb-2" />
        )}
        <EntityOwners
            entity={entity}
            className="card mb-2"
            headerClassName={headerClassName}
            titleClassName={titleClassName}
            bodyClassName={bodyClassName}
            bodyScrollableClassName={bodyScrollableClassName}
        />
        <ComponentAuthors
            catalogComponent={entity}
            className="card mb-3"
            headerClassName={headerClassName}
            titleClassName={titleClassName}
            bodyClassName={bodyClassName}
            bodyScrollableClassName={bodyScrollableClassName}
        />
        <ComponentSources
            {...props}
            catalogComponent={entity}
            className=""
            bodyScrollableClassName={bodyScrollableClassName}
        />
    </div>
)
