import classNames from 'classnames'
import React from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PackageDetailFields } from '../../../../graphql-operations'
import { CatalogEntityIcon } from '../../components/CatalogEntityIcon'

interface Props extends TelemetryProps, SettingsCascadeProps, PlatformContextProps {
    entity: PackageDetailFields
    className?: string
}

export const PackageOverviewTab: React.FunctionComponent<Props> = ({
    entity,
    className,
    telemetryService,
    settingsCascade,
    platformContext,
}) => (
    <div className={classNames('row no-gutters', className)}>
        <div className="col-md-4 col-lg-3 col-xl-2 border-right p-3">
            {entity.name && (
                <h2 className="d-flex align-items-center mb-1">
                    <CatalogEntityIcon entity={entity} className="icon-inline mr-2" /> {entity.name}
                </h2>
            )}
            {entity.description && <p className="mb-3">{entity.description}</p>}
        </div>
        <div className="col-md-8 col-lg-9 col-xl-10 p-3">hello</div>
    </div>
)
