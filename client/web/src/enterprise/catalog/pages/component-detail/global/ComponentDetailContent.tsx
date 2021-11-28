import classNames from 'classnames'
import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentDetailFields } from '../../../../../graphql-operations'
import { CatalogComponentIcon } from '../../../components/CatalogComponentIcon'

import { ComponentAuthors } from './ComponentAuthors'
import { ComponentCommits } from './ComponentCommits'
import styles from './ComponentDetailContent.module.scss'
import { ComponentSources } from './ComponentSources'

interface Props extends TelemetryProps {
    catalogComponent: CatalogComponentDetailFields
}

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ catalogComponent }) => (
    <div>
        <header>
            <h1>
                <CatalogComponentIcon catalogComponent={catalogComponent} className="icon-inline mr-1" />{' '}
                {catalogComponent.name}
            </h1>
            <ul className="list-unstyled">
                <li>
                    <strong>Owner</strong> alice
                </li>
                <li>
                    <strong>Lifecycle</strong> production
                </li>
            </ul>
        </header>
        <div className="py-4 border-top">
            <h2>Implementation</h2>
            <div className={styles.grid}>
                {/* TODO(sqs): group sources "by owner" "by tree" "by lang" etc. */}
                <ComponentSources
                    catalogComponent={catalogComponent}
                    className="card"
                    headerClassName={classNames('card-header', styles.cardHeader)}
                    titleClassName={classNames('card-title', styles.cardTitle)}
                    bodyClassName={styles.cardBody}
                />
                <div>
                    <ComponentAuthors
                        catalogComponent={catalogComponent}
                        className="card mb-3"
                        headerClassName={classNames('card-header', styles.cardHeader)}
                        titleClassName={classNames('card-title', styles.cardTitle)}
                        bodyClassName={styles.cardBody}
                    />
                    <ComponentCommits
                        catalogComponent={catalogComponent}
                        className="card"
                        headerClassName={classNames('card-header', styles.cardHeader)}
                        titleClassName={classNames('card-title', styles.cardTitle)}
                        bodyClassName={styles.cardBody}
                    />
                </div>
            </div>
        </div>
        <div className={styles.grid}>
            <section className="card card-body">
                <h3>Usage</h3>
            </section>
            <section className="card card-body">
                <h3>Depends on</h3>
            </section>
            <section className="card card-body">
                <h3>Used by</h3>
            </section>
        </div>
    </div>
)
