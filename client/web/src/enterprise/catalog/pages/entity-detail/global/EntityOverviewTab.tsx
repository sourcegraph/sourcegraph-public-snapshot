import classNames from 'classnames'
import { uniqBy } from 'lodash'
import AccountGroupIcon from 'mdi-react/AccountGroupIcon'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogEntityRelationType } from '@sourcegraph/shared/src/graphql/schema'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogEntityDetailFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { EntityGraph } from '../../../components/entity-graph/EntityGraph'

import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { EntityDetailContentCardProps } from './EntityDetailContent'
import { EntityOwners } from './EntityOwners'

interface Props extends EntityDetailContentCardProps {
    entity: CatalogEntityDetailFields
}

export const EntityOverviewTab: React.FunctionComponent<Props> = ({
    entity,
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
}) => (
    <div className="d-flex flex-column">
        {entity.__typename === 'CatalogComponent' ? (
            <>
                <div className="row2222">
                    {entity.description && <p className="mb-3">{entity.description}</p>}
                    <div className="card mb-3">
                        <div className="card-body">
                            <Link
                                to={`/search?q=context:c/${entity.name}`}
                                className="d-flex align-items-center btn btn-outline-secondary"
                            >
                                <SearchIcon className="icon-inline" /> Search in {entity.name}...
                            </Link>
                        </div>
                        <ComponentSourceDefinitions
                            catalogComponent={entity}
                            listGroupClassName="list-group-flush"
                            className="border-top"
                        />
                        {entity.commits?.nodes[0] && (
                            <LastCommit commit={entity.commits.nodes[0]} className="card-footer" />
                        )}
                    </div>

                    {/* owner-docs-API def -- authorities. then who you could ask. */}
                    <div className="d-flex">
                        <Link to="#" className="d-flex align-items-center text-body mb-3 mr-3 p-2">
                            <FileDocumentIcon className="icon-inline mr-2" />
                            Documentation
                        </Link>
                        <Link to="#" className="d-flex align-items-center text-body mb-3 mr-3 p-2">
                            <SourcePullIcon className="icon-inline mr-2" />
                            Pull requests
                        </Link>
                        <Link to="#" className="d-flex align-items-center text-body mb-3 p-2">
                            <AlertCircleOutlineIcon className="icon-inline mr-2" />
                            Issues
                        </Link>
                    </div>
                    <hr className="my-3" />
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <AccountGroupIcon className="icon-inline mr-2 d-none" />
                        <span className="badge badge-primary bg-transparent border-primary border text-primary mr-2">
                            Owners
                        </span>
                        <small>
                            @unknwon <span className="text-muted">50%</span> &nbsp;@tsenart{' '}
                            <span className="text-muted">42%</span> &nbsp; <span className="text-muted">3 more</span>
                        </small>
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <AccountGroupIcon className="icon-inline mr-2 d-none" />
                        <span className="badge badge-success bg-transparent border-primary border text-primary mr-2">
                            Authors
                        </span>
                        <small>
                            Joe Chen <span className="text-muted">50%</span> &nbsp;Tomas Senart{' '}
                            <span className="text-muted">31%</span> &nbsp; <span className="text-muted">3 more</span>
                        </small>
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body">
                        <AccountGroupIcon className="icon-inline mr-2 d-none" />
                        <span className="badge badge-success bg-transparent border-primary border text-primary mr-2">
                            Users
                        </span>
                        <small>
                            Alice Smith <span className="text-muted">13</span> &nbsp;Michael McDonald{' '}
                            <span className="text-muted">6</span> &nbsp; <span className="text-muted">3 more</span>
                        </small>
                    </Link>
                    {false && (
                        <EntityOwners
                            entity={entity}
                            className="card mb-2"
                            headerClassName={headerClassName}
                            titleClassName={titleClassName}
                            bodyClassName={bodyClassName}
                            bodyScrollableClassName={bodyScrollableClassName}
                        />
                    )}
                </div>

                <EntityGraph
                    graph={{
                        edges: entity.relatedEntities.edges.map(edge =>
                            edge.type === CatalogEntityRelationType.DEPENDS_ON
                                ? {
                                      type: edge.type,
                                      outNode: entity,
                                      inNode: edge.node,
                                  }
                                : {
                                      type: CatalogEntityRelationType.DEPENDS_ON,
                                      outNode: edge.node,
                                      inNode: entity,
                                  }
                        ),
                        nodes: uniqBy(entity.relatedEntities.edges.map(edge => edge.node).concat(entity), 'id'),
                    }}
                    activeNodeID={entity.id}
                />
            </>
        ) : (
            <div>Typename is {entity.__typename}</div>
        )}
    </div>
)

const LastCommit: React.FunctionComponent<{
    commit: NonNullable<CatalogEntityDetailFields['commits']>['nodes'][0]
    className?: string
}> = ({ commit, className }) => (
    <div className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={commit.author.person} size={14} />
        <PersonLink person={commit.author.person} className="font-weight-bold mr-2 flex-shrink-0" />
        <Link to={commit.url} className="text-truncate flex-grow-1 text-body mr-2" title={commit.message}>
            {commit.subject}
        </Link>
        <small className="text-nowrap text-muted">
            <Link to={commit.url} className="text-monospace text-muted mr-2 d-none d-md-inline">
                {commit.abbreviatedOID}
            </Link>
            <Timestamp date={commit.author.date} noAbout={true} />
        </small>
    </div>
)
