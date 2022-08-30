import React from 'react'

import { mdiAccount, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Link, H3, Text, Tooltip, Badge } from '@sourcegraph/wildcard'

import { ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'

import styles from './ExternalServiceCard.module.scss'

interface ExternalServiceCardProps {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription?: string

    kind: ExternalServiceKind

    namespace?: ExternalServiceFields['namespace']

    to?: string
    className?: string
    enabled?: boolean
    requiredTier?: string
}

export const ExternalServiceCard: React.FunctionComponent<React.PropsWithChildren<ExternalServiceCardProps>> = ({
    title,
    icon: CardIcon,
    shortDescription,
    to,
    kind,
    namespace,
    className = '',
    enabled = true,
    requiredTier = '',
}) => {
    let cardTitle = (<H3 className={shortDescription ? 'mb-0' : 'mt-1 mb-0'}>
                    {title}
                    {namespace && (
                        <small>
                            {' '}
                            by
                            <Icon aria-hidden={true} svgPath={mdiAccount} />
                            <Link to={namespace.url}>{namespace.namespaceName}</Link>
                        </small>
                    )}
                </H3>)
    cardTitle = !enabled && requiredTier ? (
        <Tooltip content="Test">
            {cardTitle}
        </Tooltip>
        ) : cardTitle
    const children = (
        <div className={classNames('p-3 d-flex align-items-start border' + (enabled ? '' : ' text-muted'), className)}>
            <Icon disabled={!enabled} className={classNames('mb-0 mr-3', styles.icon)} as={CardIcon} aria-hidden={true} />
            <div className="flex-1">
                {cardTitle}
                {shortDescription && <Text className="mb-0 text-muted">{shortDescription}</Text>}
            </div>
            {to && enabled && <Icon className="align-self-center" svgPath={mdiChevronRight} inline={false} aria-hidden={true} />}
            {!enabled && <Badge className="align-self-center" variant="outlineSecondary">{requiredTier.toUpperCase()}</Badge>}
        </div>
    )
    return to && enabled ? (
        <Link
            className="d-block text-left text-body text-decoration-none"
            to={to}
            data-test-external-service-card-link={kind}
        >
            {children}
        </Link>
    ) : (
        children
    )
}
