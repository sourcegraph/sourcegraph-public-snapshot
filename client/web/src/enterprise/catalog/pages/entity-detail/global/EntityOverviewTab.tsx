import { uniqBy } from 'lodash'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogEntityRelationType } from '@sourcegraph/shared/src/graphql/schema'

import { CatalogEntityDetailFields } from '../../../../../graphql-operations'
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
                <Link
                    to={`/search?q=context:c/${entity.name}`}
                    className="d-flex align-items-center mb-2 btn btn-outline-secondary"
                >
                    <SearchIcon className="icon-inline" /> Search in {entity.name}...
                </Link>
                <ComponentSourceDefinitions catalogComponent={entity} className="mb-2" />
                <EntityOwners
                    entity={entity}
                    className="card mb-2"
                    headerClassName={headerClassName}
                    titleClassName={titleClassName}
                    bodyClassName={bodyClassName}
                    bodyScrollableClassName={bodyScrollableClassName}
                />
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
