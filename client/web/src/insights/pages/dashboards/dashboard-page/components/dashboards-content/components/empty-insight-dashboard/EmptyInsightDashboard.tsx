import classnames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import styles from './EmptyInsightDashboard.module.scss'

export const EmptyInsightDashboard: React.FunctionComponent = () => (
    <div>
        <section className={styles.emptySection}>
            <Link to="/insights/create" className={classnames(styles.itemCard, 'card')}>
                <PlusIcon size="2rem" />
                <span>Create new Insight</span>
            </Link>
            <span className="d-flex justify-content-center mt-3">
                <span>
                    ...or add existing Insights from <Link to="/insights/dashboards/all">All Insights</Link>
                </span>
            </span>
        </section>
    </div>
)
