import classNames from 'classnames'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { useLocation } from 'react-router'
import { Link } from 'react-router-dom'
import { of } from 'rxjs'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { useQuery, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SourceLocationSetUsageResult, SourceLocationSetUsageVariables } from '../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { fetchHighlightedFileLineRanges } from '../../../../repo/backend'
import { ComponentIcon } from '../../components/ComponentIcon'

import { PersonList } from './PersonList'

interface Props extends SettingsCascadeProps, TelemetryProps {
    sourceLocationSet: Scalars['ID']
    className?: string
}

const SOURCE_LOCATION_SET_USAGE = gql`
    query SourceLocationSetUsage($node: ID!) {
        node(id: $node) {
            ...SourceLocationSetUsageFields
        }
    }

    fragment SourceLocationSetUsageFields on SourceLocationSet {
        __typename
        id
        usage {
            locations {
                nodes {
                    range {
                        start {
                            line
                            character
                        }
                        end {
                            line
                            character
                        }
                    }
                    resource {
                        path
                        commit {
                            oid
                        }
                        repository {
                            name
                        }
                    }
                }
            }
            people {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                authoredLineCount
                lastCommit {
                    author {
                        date
                    }
                }
            }
            components {
                node {
                    __typename
                    id
                    name
                    kind
                    url
                }
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const UsageTab: React.FunctionComponent<Props> = ({
    sourceLocationSet,
    className,
    settingsCascade,
    telemetryService,
}) => {
    const { data, error, loading } = useQuery<SourceLocationSetUsageResult, SourceLocationSetUsageVariables>(
        SOURCE_LOCATION_SET_USAGE,
        {
            variables: { node: sourceLocationSet },
            fetchPolicy: 'cache-first',
        }
    )

    const location = useLocation()

    if (loading && !data) {
        return <LoadingSpinner />
    }
    if (error && !data) {
        return <ErrorAlert error={error} />
    }
    if (!data || !data.node || !('id' in data.node)) {
        return <ErrorAlert error="TODO(sqs)" />
    }
    if (!data.node.usage) {
        return <ErrorAlert error="No usage information" />
    }

    const { people: peopleEdges, components: componentEdges, locations } = data.node.usage
    return locations && locations.nodes.length > 0 ? (
        <div className={classNames('row no-gutters', className)}>
            <div className="col-md-8 col-lg-9 col-xl-10 p-3">
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
                    parentContainerIsEmpty={false}
                    telemetryService={telemetryService}
                />
            </div>
            <div className="col-md-4 col-lg-3 col-xl-2 border-left p-3">
                <h4 className="font-weight-bold">From components</h4>
                <ol className={classNames('list-group mb-3')}>
                    {componentEdges.map(edge => (
                        <li key={edge.node.id} className={classNames('list-group-item')}>
                            <Link to={edge.node.url} className="d-flex align-items-center text-body">
                                <ComponentIcon component={edge.node} className="icon-inline text-muted mr-2" />
                                {edge.node.name}
                            </Link>
                        </li>
                    ))}
                </ol>
                <PersonList
                    title="Callers"
                    listTag="ol"
                    orientation="vertical"
                    items={
                        peopleEdges
                            ? peopleEdges.map(({ node: person, authoredLineCount, lastCommit }) => ({
                                  person,
                                  text: `${authoredLineCount} ${pluralize('use', authoredLineCount)}`,
                                  date: lastCommit.author.date,
                              }))
                            : []
                    }
                    className="mb-3"
                />
            </div>
        </div>
    ) : (
        <p>No uses found</p>
    )
}
