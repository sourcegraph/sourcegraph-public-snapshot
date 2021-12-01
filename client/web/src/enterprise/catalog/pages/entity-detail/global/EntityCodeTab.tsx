import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentCommits } from './ComponentCommits'
import { ComponentSources } from './ComponentSources'
import { EntityDetailContentCardProps } from './EntityDetailContent'

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
        {/* TODO(sqs): group sources "by owner" "by tree" "by lang" etc. */}
        <ComponentSources
            {...props}
            catalogComponent={entity}
            className=""
            bodyScrollableClassName={bodyScrollableClassName}
        />
        <div className="d-flex flex-column">
            <ComponentAuthors
                catalogComponent={entity}
                className="card mb-3"
                headerClassName={headerClassName}
                titleClassName={titleClassName}
                bodyClassName={bodyClassName}
                bodyScrollableClassName={bodyScrollableClassName}
            />
            <ComponentCommits
                catalogComponent={entity}
                className="card overflow-hidden"
                headerClassName={headerClassName}
                titleClassName={titleClassName}
                bodyClassName={bodyClassName}
                bodyScrollableClassName={bodyScrollableClassName}
            />
        </div>
        {/* TODO(sqs): add "Depends on" */}
    </div>
)
