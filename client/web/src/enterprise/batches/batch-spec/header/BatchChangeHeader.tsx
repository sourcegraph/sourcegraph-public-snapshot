import React from 'react'

import classNames from 'classnames'
import { compact } from 'lodash'

import { PageHeader, FeedbackBadge } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../../batches/icons'

import styles from './BatchChangeHeader.module.scss'

interface BatchChangeHeaderProps {
    className?: string
    /** The namespace to display in the `PageHeader`. */
    namespace?: { to: string; text: string }
    /** The secondary title to use in the `PageHeader`, after the namespace. */
    title: { to?: string; text: string }
    /** The description to use in the `PageHeader`, under the namespace and title row. */
    description?: React.ReactNode
}

export const BatchChangeHeader: React.FC<BatchChangeHeaderProps> = ({ className, namespace, title, description }) => (
    <PageHeader
        path={compact([{ icon: BatchChangesIcon }, namespace, title])}
        className={classNames('flex-1 pb-2', styles.header, className)}
        description={
            description || 'Run custom code over hundreds of repositories and manage the resulting changesets.'
        }
        annotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
    />
)
