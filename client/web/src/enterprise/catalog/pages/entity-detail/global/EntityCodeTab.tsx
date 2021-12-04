import classNames from 'classnames'
import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { ComponentSources } from './ComponentSources'
import { EntityCodeOwners } from './EntityCodeOwners'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: CatalogEntityDetailFields
    className?: string
}

export const EntityCodeTab: React.FunctionComponent<Props> = ({ entity, className, ...props }) => (
    <div className={classNames('container p-3', className)}>
        {entity.__typename === 'CatalogComponent' && (
            <ComponentSourceDefinitions catalogComponent={entity} className="mb-2" />
        )}
        <EntityCodeOwners entity={entity} className="card mb-2" />
        <ComponentAuthors catalogComponent={entity} className="card mb-3 overflow-auto" />
        <ComponentSources {...props} catalogComponent={entity} />
    </div>
)
