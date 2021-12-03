import classNames from 'classnames'
import SettingsIcon from 'mdi-react/SettingsIcon'
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
import { EntityOwner } from '../../../components/entity-owner/EntityOwner'

import { ComponentAPI } from './ComponentApi'
import { ComponentDocumentation } from './ComponentDocumentation'
import { ComponentUsage } from './ComponentUsage'
import { EntityChangesTab } from './EntityChangesTab'
import { EntityCodeTab } from './EntityCodeTab'
import styles from './EntityDetailContent.module.scss'
import { EntityOverviewTab } from './EntityOverviewTab'
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

const cardProps: EntityDetailContentCardProps = {
    headerClassName: classNames('card-header', styles.cardHeader),
    titleClassName: classNames('card-title', styles.cardTitle),
    bodyClassName: classNames('card-body', styles.cardBody),
    bodyScrollableClassName: styles.cardBodyScrollable,
}

export const EntityDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof TabRouter>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    label: 'Overview',
                    element: <EntityOverviewTab {...cardProps} entity={entity} />,
                },
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'code',
                          label: 'Code',
                          element: <EntityCodeTab {...props} {...cardProps} entity={entity} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'changes',
                          label: 'Changes',
                          element: <EntityChangesTab {...props} {...cardProps} entity={entity} />,
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'docs',
                          label: 'Docs',
                          element: <ComponentDocumentation catalogComponent={entity} />,
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'api',
                          label: 'API',
                          element: <ComponentAPI {...props} catalogComponent={entity} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'usage',
                          label: 'Usage',
                          element: <ComponentUsage {...props} {...cardProps} catalogComponent={entity} />,
                      }
                    : null,
                false
                    ? {
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
                      }
                    : null,
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
                actions={
                    // eslint-disable-next-line react/forbid-dom-props
                    <nav className="d-flex align-items-center" style={{ marginBottom: '-5px' }}>
                        <div className="d-inline-block mr-4">
                            <span className="small font-weight-bold">Lifecycle</span>
                            <br />
                            {entity.lifecycle.toLowerCase()}
                        </div>
                        <div className="text-body d-inline-block mr-4">
                            <span className="small font-weight-bold">Owner</span>
                            <br />
                            <EntityOwner owner={entity.owner} />
                        </div>
                        <Link to="#" className="d-inline-block btn btn-secondary btn-sm p-2 mb-0">
                            <SettingsIcon className="icon-inline" />
                        </Link>
                    </nav>
                }
                className="mt-3 mb-1"
            />
            <div className="mb-4" />
            <TabRouter tabs={tabs} />
        </div>
    )
}
