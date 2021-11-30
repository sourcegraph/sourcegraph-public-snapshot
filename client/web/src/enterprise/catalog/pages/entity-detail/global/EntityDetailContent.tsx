import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogEntityDetailFields } from '../../../../../graphql-operations'
import { catalogEntityIconComponent } from '../../../components/CatalogEntityIcon'
import { EntityGraph } from '../../../components/entity-graph/EntityGraph'

import { ComponentAPI } from './ComponentApi'
import { ComponentAuthors } from './ComponentAuthors'
import { ComponentCommits } from './ComponentCommits'
import styles from './ComponentDetailContent.module.scss'
import { ComponentDocumentation } from './ComponentDocumentation'
import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { ComponentSources } from './ComponentSources'
import { ComponentUsage } from './ComponentUsage'
import { TabRouter } from './TabRouter'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: CatalogEntityDetailFields
}

export interface EntityDetailContentCardProps {
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
    bodyScrollableClassName?: string
}

export const EntityDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof TabRouter>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    label: 'Overview',
                    element: (
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
                                    <EntityGraph
                                        graph={{
                                            edges: entity.relatedEntities.edges.map(edge =>
                                                edge.type === 'DEPENDS_ON'
                                                    ? {
                                                          outNode: edge.node,
                                                          outType: edge.type,
                                                          inNode: entity,
                                                          inType: edge.type,
                                                      }
                                                    : {
                                                          outNode: entity,
                                                          outType: edge.type,
                                                          inNode: edge.node,
                                                          inType: edge.type,
                                                      }
                                            ),
                                            nodes: entity.relatedEntities.edges.map(edge => edge.node).concat(entity),
                                        }}
                                        activeNodeID={entity.id}
                                    />
                                    {false && (
                                        <>
                                            <ComponentAuthors
                                                catalogComponent={entity}
                                                className="card mb-3"
                                                headerClassName={classNames('card-header', styles.cardHeader)}
                                                titleClassName={classNames('card-title', styles.cardTitle)}
                                                bodyClassName={styles.cardBody}
                                                bodyScrollableClassName={styles.cardBodyScrollable}
                                            />
                                            <ComponentCommits
                                                catalogComponent={entity}
                                                className="card overflow-hidden"
                                                headerClassName={classNames('card-header', styles.cardHeader)}
                                                titleClassName={classNames('card-title', styles.cardTitle)}
                                                bodyClassName={styles.cardBody}
                                                bodyScrollableClassName={styles.cardBodyScrollable}
                                            />
                                        </>
                                    )}
                                </>
                            ) : (
                                <div>Typename is {entity.__typename}</div>
                            )}
                        </div>
                    ),
                },
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'code',
                          label: 'Code',
                          element: (
                              <div className={styles.grid}>
                                  {/* TODO(sqs): group sources "by owner" "by tree" "by lang" etc. */}
                                  <ComponentSources
                                      {...props}
                                      catalogComponent={entity}
                                      className=""
                                      bodyScrollableClassName={styles.cardBodyScrollable}
                                  />
                                  <div className="d-flex flex-column">
                                      <ComponentAuthors
                                          catalogComponent={entity}
                                          className="card mb-3"
                                          headerClassName={classNames('card-header', styles.cardHeader)}
                                          titleClassName={classNames('card-title', styles.cardTitle)}
                                          bodyClassName={styles.cardBody}
                                          bodyScrollableClassName={styles.cardBodyScrollable}
                                      />
                                      <ComponentCommits
                                          catalogComponent={entity}
                                          className="card overflow-hidden"
                                          headerClassName={classNames('card-header', styles.cardHeader)}
                                          titleClassName={classNames('card-title', styles.cardTitle)}
                                          bodyClassName={styles.cardBody}
                                          bodyScrollableClassName={styles.cardBodyScrollable}
                                      />
                                  </div>
                                  {/* TODO(sqs): add "Depends on" */}
                              </div>
                          ),
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'docs',
                          label: 'Docs',
                          element: <ComponentDocumentation catalogComponent={entity} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'api',
                          label: 'API',
                          element: <ComponentAPI {...props} catalogComponent={entity} className="" />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'usage',
                          label: 'Usage',
                          element: (
                              <ComponentUsage
                                  {...props}
                                  catalogComponent={entity}
                                  className=""
                                  bodyClassName={styles.cardBody}
                                  bodyScrollableClassName={styles.cardBodyScrollable}
                              />
                          ),
                      }
                    : null,
                {
                    path: 'spec',
                    label: 'Spec',
                    element: (
                        <>
                            <p>
                                Edit the JSON specification for this {entity.__typename.toLowerCase()} in source
                                control.
                            </p>
                            <Container>
                                <pre>
                                    <code>
                                        {JSON.stringify(
                                            {
                                                name: entity.name,
                                                type: entity.type,
                                                description: entity.description,
                                                kind: 'kind' in entity ? entity.kind : undefined,
                                                sourceLocations:
                                                    'sourceLocations' in entity
                                                        ? entity.sourceLocations.map(location_ => ({
                                                              repo: location_.repository.name,
                                                              path: location_.path,
                                                          }))
                                                        : undefined,
                                            },
                                            null,
                                            2
                                        )}
                                    </code>
                                </pre>
                            </Container>
                        </>
                    ),
                },
            ].filter(isDefined),
        [entity, props]
    )
    return (
        <div>
            <PageHeader
                path={[
                    { icon: CatalogIcon, to: '/catalog' },
                    {
                        icon: catalogEntityIconComponent(entity),
                        text: entity.name,
                    },
                ]}
                className="mb-1"
            />
            {entity.description && <p className="mb-0 text-muted">{entity.description}</p>}
            <div className="mb-4" />
            <TabRouter tabs={tabs} />
        </div>
    )
}
