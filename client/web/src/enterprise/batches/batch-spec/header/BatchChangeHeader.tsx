import React from 'react'

import classNames from 'classnames'

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
        className={classNames('flex-1 pb-2', styles.header, className)}
        description={description || 'Run and manage large-scale changes across many repositories.'}
        annotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
    >
        <PageHeader.Heading as="h2" styleAs="h1">
            <PageHeader.Breadcrumb icon={BatchChangesIcon} />
            {namespace && <PageHeader.Breadcrumb to={namespace.to}>{namespace.text}</PageHeader.Breadcrumb>}
            <PageHeader.Breadcrumb to={title.to}>{title.text}</PageHeader.Breadcrumb>
        </PageHeader.Heading>
    </PageHeader>
)
