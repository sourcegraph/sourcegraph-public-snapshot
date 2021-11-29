import classNames from 'classnames'
import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogComponentDetailFields } from '../../../../../graphql-operations'
import { CATALOG_COMPONENT_ICON_BY_KIND } from '../../../components/CatalogComponentIcon'

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentCommits } from './ComponentCommits'
import styles from './ComponentDetailContent.module.scss'
import { ComponentSources } from './ComponentSources'
import { ComponentUsage } from './ComponentUsage'
import { TabRouter } from './TabRouter'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    catalogComponent: CatalogComponentDetailFields
}

export interface ComponentDetailContentCardProps {
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
    bodyScrollableClassName?: string
}

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ catalogComponent, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof TabRouter>['tabs']>(
        () => [
            {
                path: '',
                exact: true,
                label: 'Documentation',
                element: <p>Documentation</p>,
            },
            {
                path: 'impl',
                label: 'Implementation',
                element: (
                    <div className={styles.grid}>
                        {/* TODO(sqs): group sources "by owner" "by tree" "by lang" etc. */}
                        <ComponentSources
                            {...props}
                            catalogComponent={catalogComponent}
                            className=""
                            bodyScrollableClassName={styles.cardBodyScrollable}
                        />
                        <div className="d-flex flex-column">
                            <ComponentAuthors
                                catalogComponent={catalogComponent}
                                className="card mb-3"
                                headerClassName={classNames('card-header', styles.cardHeader)}
                                titleClassName={classNames('card-title', styles.cardTitle)}
                                bodyClassName={styles.cardBody}
                                bodyScrollableClassName={styles.cardBodyScrollable}
                            />
                            <ComponentCommits
                                catalogComponent={catalogComponent}
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
            },
            {
                path: 'api',
                label: 'API',
                element: <p>API</p>,
            },
            {
                path: 'usage',
                label: 'Usage',
                element: (
                    <ComponentUsage
                        {...props}
                        catalogComponent={catalogComponent}
                        className="card"
                        headerClassName={classNames('card-header', styles.cardHeader)}
                        titleClassName={classNames('card-title', styles.cardTitle)}
                        bodyClassName={styles.cardBody}
                        bodyScrollableClassName={styles.cardBodyScrollable}
                    />
                ),
            },
            {
                path: 'spec',
                label: 'Spec',
                element: <p>Spec</p>,
            },
        ],
        [catalogComponent, props]
    )
    return (
        <div>
            <PageHeader
                path={[
                    { icon: CatalogIcon, to: '/catalog' },
                    {
                        icon: CATALOG_COMPONENT_ICON_BY_KIND[catalogComponent.kind],
                        text: catalogComponent.name,
                    },
                ]}
                className="mb-3"
            />
            <ul className="list-unstyled">
                <li>
                    <strong>Owner</strong> alice
                </li>
                <li>
                    <strong>Lifecycle</strong> production
                </li>
            </ul>
            <TabRouter tabs={tabs} />
        </div>
    )
}
