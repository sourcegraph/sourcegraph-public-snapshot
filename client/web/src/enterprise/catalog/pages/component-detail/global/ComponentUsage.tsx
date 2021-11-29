import classNames from 'classnames'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { useLocation } from 'react-router'
import { of } from 'rxjs'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogComponentUsageFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { fetchHighlightedFileLineRanges } from '../../../../../repo/backend'
import { UserAvatar } from '../../../../../user/UserAvatar'

import { ComponentDetailContentCardProps } from './ComponentDetailContent'
import styles from './ComponentUsage.module.scss'

interface Props
    extends Pick<ComponentDetailContentCardProps, 'className' | 'bodyClassName' | 'bodyScrollableClassName'>,
        SettingsCascadeProps,
        TelemetryProps {
    catalogComponent: CatalogComponentUsageFields
}

export const ComponentUsage: React.FunctionComponent<Props> = ({
    catalogComponent: { usage },
    className,
    bodyClassName,
    bodyScrollableClassName,
    settingsCascade,
    telemetryService,
}) => {
    const { data, error, loading } = useQuery<CatalogComponentByNameResult, CatalogComponentByNameVariables>(
        CATALOG_COMPONENT_BY_NAME,
        {
            variables: { name: catalogComponentName },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    const location = useLocation()

    if (!usage) {
        return (
            <div className={className}>
                <div className="alert-warning">Unable to determine usage information (no usage patterns specified)</div>
            </div>
        )
    }

    const { locations, callers } = usage
    return locations && locations.nodes.length > 0 ? (
        <div className={className}>
            <ol className={classNames('list-group list-group-horizontal overflow-auto flex-shrink-0', bodyClassName)}>
                {callers.map(caller => (
                    <li
                        key={caller.person.email}
                        className={classNames('list-group-item text-center pt-2', styles.author)}
                    >
                        <div>
                            <UserAvatar className="icon-inline" user={caller.person} />
                        </div>
                        <PersonLink person={caller.person} className="text-muted small text-truncate d-block" />
                        <div className={classNames(styles.lineCount)}>
                            {caller.authoredLineCount} {pluralize('use', caller.authoredLineCount)}
                        </div>
                        <div className={classNames('text-muted', styles.lastCommit)}>
                            <Timestamp date={caller.lastCommit.author.date} noAbout={true} />
                        </div>
                    </li>
                ))}
            </ol>
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
