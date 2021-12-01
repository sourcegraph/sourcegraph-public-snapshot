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

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentCommits } from './ComponentCommits'
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
                        <EntityOwners
                            entity={entity}
                            className="card mb-2"
                            headerClassName={headerClassName}
                            titleClassName={titleClassName}
                            bodyClassName={bodyClassName}
                            bodyScrollableClassName={bodyScrollableClassName}
                        />
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
                {false && (
                    <>
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
                    </>
                )}
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
