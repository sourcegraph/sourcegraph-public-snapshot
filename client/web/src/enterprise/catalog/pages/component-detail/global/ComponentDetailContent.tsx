import classNames from 'classnames'
import React from 'react'

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

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ catalogComponent, ...props }) => (
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
            description={
                <ul className="list-unstyled">
                    <li>
                        <strong>Owner</strong> alice
                    </li>
                    <li>
                        <strong>Lifecycle</strong> production
                    </li>
                </ul>
            }
        />
        {true && (
            <div className="py-4 border-top">
                <h2>Implementation</h2>
                <div className={styles.grid}>
                    {/* TODO(sqs): group sources "by owner" "by tree" "by lang" etc. */}
                    <ComponentSources
                        {...props}
                        catalogComponent={catalogComponent}
                        className="card"
                        headerClassName={classNames('card-header', styles.cardHeader)}
                        titleClassName={classNames('card-title', styles.cardTitle)}
                        bodyClassName={styles.cardBody}
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
                </div>
                {/* TODO(sqs): add "Depends on" */}
            </div>
        )}
        <div className="py-4 border-top">
            <h2>Usage</h2>
            <div className={styles.grid}>
                <ComponentUsage
                    {...props}
                    catalogComponent={catalogComponent}
                    className="card"
                    headerClassName={classNames('card-header', styles.cardHeader)}
                    titleClassName={classNames('card-title', styles.cardTitle)}
                    bodyClassName={styles.cardBody}
                    bodyScrollableClassName={styles.cardBodyScrollable}
                />
            </div>
        </div>
    </div>
)
