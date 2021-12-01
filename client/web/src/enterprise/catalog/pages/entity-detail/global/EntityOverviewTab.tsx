import classNames from 'classnames'
import { uniqBy } from 'lodash'
import SearchIcon from 'mdi-react/SearchIcon'
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
                <div className="row">
                    <div className="col-md-7">
                        <div className="card">
                            <div className="card-body">
                                <Link
                                    to={`/search?q=context:c/${entity.name}`}
                                    className="d-flex align-items-center btn btn-outline-secondary"
                                >
                                    <SearchIcon className="icon-inline" /> Search in {entity.name}...
                                </Link>
                            </div>
                            {entity.commits?.nodes[0] && (
                                <LastCommit commit={entity.commits.nodes[0]} className="card-footer" />
                            )}
                            <ComponentSourceDefinitions
                                catalogComponent={entity}
                                listGroupClassName="list-group-flush"
                                className="border-top"
                            />
                        </div>
                    </div>
                    <div className="col-md-5">
                        {/* owner-docs-API def -- authorities. then who you could ask. */}
                        <div className="card">
                            <div className="d-flex justify-content-between">
                                <Link to="#" className="btn btn-lg btn-outline-secondary flex-grow-1 rounded-0">
                                    Docs
                                </Link>
                                <Link to="#" className="btn btn-lg btn-outline-secondary flex-grow-1 rounded-0">
                                    API
                                </Link>
                                <Link to="#" className="btn btn-lg btn-outline-secondary flex-grow-1 rounded-0">
                                    Owner
                                </Link>
                            </div>
                            <p className="card-body border-top mb-0">
                                <strong>Authors</strong>&nbsp;{' '}
                                <small>
                                    @ziyang <span className="text-muted">81%</span> &nbsp;@fatima{' '}
                                    <span className="text-muted">15%</span> &nbsp;@walter{' '}
                                    <span className="text-muted">12%</span> &nbsp;
                                </small>
                            </p>
                            <p className="card-body border-top mb-0">
                                <strong>Callers</strong>&nbsp;{' '}
                                <small>
                                    @alice <span className="text-muted">51</span> &nbsp;@bob{' '}
                                    <span className="text-muted">31</span> &nbsp;
                                </small>
                            </p>
                        </div>
                        {false && (
                            <div className="card">
                                <p className="card-body mb-0">
                                    <strong>Owners</strong>&nbsp;{' '}
                                    <small>
                                        @unknwon <span className="text-muted">50%</span> &nbsp;@tsenart{' '}
                                        <span className="text-muted">42%</span> &nbsp;
                                    </small>
                                </p>
                                <p className="card-body border-top mb-0">
                                    <strong>Authors</strong>&nbsp;{' '}
                                    <small>
                                        @ziyang <span className="text-muted">81%</span> &nbsp;@fatima{' '}
                                        <span className="text-muted">15%</span> &nbsp;@walter{' '}
                                        <span className="text-muted">12%</span> &nbsp;
                                    </small>
                                </p>
                                <p className="card-body border-top mb-0">
                                    <strong>Callers</strong>&nbsp;{' '}
                                    <small>
                                        @alice <span className="text-muted">51</span> &nbsp;@bob{' '}
                                        <span className="text-muted">31</span> &nbsp;
                                    </small>
                                </p>
                            </div>
                        )}
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
            <Timestamp date={commit.author.date} noAbout={true} />
        </small>
    </div>
)
