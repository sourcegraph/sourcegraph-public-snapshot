import classNames from 'classnames'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { useLocation } from 'react-router'
import { of } from 'rxjs'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { CatalogComponentUsageFields } from '../../../../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../../../../repo/backend'

import { ComponentDetailContentCardProps } from './ComponentDetailContent'

interface Props extends ComponentDetailContentCardProps, SettingsCascadeProps, TelemetryProps {
    catalogComponent: CatalogComponentUsageFields
}

export const ComponentUsage: React.FunctionComponent<Props> = ({
    catalogComponent: {
        usage: { locations },
    },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
    settingsCascade,
    telemetryService,
}) => {
    const location = useLocation()

    return locations && locations.nodes.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Locations</h3>
            </header>
            <style>
                {
                    'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container { border: solid 1px var(--border-color) !important; border-left: none !important; border-right: none !important; margin: 0; } .result-container small { display: none; } .result-container__header > .mdi-icon { display: none; } .result-container__header-divider { display: none; } .result-container__header { padding-left: 0.25rem; } .FileMatchChildren-module__file-match-children { border: none !important; } .result-container { border: none !important; }'
                }
            </style>
            <FileLocations
                location={location}
                locations={of(
                    locations.nodes.map<Location>(location => ({
                        uri: makeRepoURI({
                            repoName: location.resource.repository.name,
                            commitID: location.resource.commit.oid,
                            filePath: location.resource.path,
                        }),
                        range: location.range!,
                    }))
                )}
                icon={SourceRepositoryIcon}
                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                settingsCascade={settingsCascade}
                className={classNames(bodyClassName, bodyScrollableClassName)}
                parentContainerIsEmpty={false}
                telemetryService={telemetryService}
            />
        </div>
    ) : (
        <p>No uses found</p>
    )
}
