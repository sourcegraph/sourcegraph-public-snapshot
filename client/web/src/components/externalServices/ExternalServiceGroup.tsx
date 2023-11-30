import { type FC, useState } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, Link, ProductStatusBadge, Badge } from '@sourcegraph/wildcard'

import type { AddExternalServiceOptions } from './externalServices'

import styles from './ExternalServiceGroup.module.scss'

interface ExternalServiceGroupProps {
    name: string
    services: AddExternalServiceOptionsWithID[]
    description?: string
    renderIcon: boolean

    icon?: React.ComponentType<{ className?: string }>
}

export interface AddExternalServiceOptionsWithID extends AddExternalServiceOptions {
    serviceID: string
    enabled?: boolean
    badge?: string
    tooltip?: string
}

export const ExternalServiceGroup: FC<ExternalServiceGroupProps> = ({
    name,
    services,
    description,
    icon,
    renderIcon,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(true)
    const toggleIsOpen = (): void => setIsOpen(prevIsOpen => !prevIsOpen)

    if (services.length === 1) {
        const [service] = services
        return (
            <div
                className={classNames('mb-3 px-2 border', styles.externalServiceGroupNode, {
                    [styles.externalServiceGroupEnabledNode]: service.enabled,
                })}
            >
                <ExternalServiceGroupNode service={service} renderIcon={true} />
            </div>
        )
    }

    return (
        <div className={classNames('border', styles.externalServiceGroupContainer)}>
            <Button
                className={classNames(styles.externalServiceGroupHeader, {
                    'border-bottom': isOpen,
                })}
                onClick={toggleIsOpen}
            >
                <div>
                    {icon && <Icon className="mb-0 mr-1" as={icon} aria-hidden={true} />} {name}
                    {'  '}
                    {description && <small className={styles.externalServiceGroupDescription}>{description}</small>}
                </div>

                <Icon aria-hidden={true} svgPath={isOpen ? mdiChevronUp : mdiChevronDown} />
            </Button>
            {isOpen && (
                <ul className={styles.externalServiceGroupBody}>
                    {services.map((service, index) => (
                        <li
                            key={index}
                            className={classNames(styles.externalServiceGroupNode, {
                                [styles.externalServiceGroupEnabledNode]: service.enabled,
                            })}
                        >
                            <ExternalServiceGroupNode service={service} renderIcon={renderIcon} />
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}

interface ExternalServiceGroupNodeProps {
    service: AddExternalServiceOptionsWithID
    renderIcon: boolean
}

const ExternalServiceGroupNode: FC<ExternalServiceGroupNodeProps> = ({ service, renderIcon }) => {
    const isServiceEnabled = service.enabled
    const children = (
        <div
            className={classNames(styles.externalServiceGroupNodeWrapper, {
                'text-muted': !isServiceEnabled,
                'py-2': !isServiceEnabled,
            })}
        >
            {renderIcon && <Icon inline={true} className="mb-0 mr-1" as={service.icon} aria-hidden={true} />}
            <div className={styles.externalServiceGroupNodeDisplayName}>
                <span>{service.title}</span>
                {'  '}
                {service.status && <ProductStatusBadge status={service.status} className="mx-1" />}
                {service.badge && (
                    <Badge className="mx-1" variant="outlineSecondary">
                        {service.badge.toUpperCase()}
                    </Badge>
                )}
                <span
                    className={classNames(styles.externalServiceGroupNodeDescription, {
                        'd-block': Boolean(service.status || service.badge),
                    })}
                >
                    {service.shortDescription}
                </span>
            </div>
        </div>
    )

    return service.enabled ? (
        <Link
            className={classNames(styles.externalServiceGroupLink, 'text-left text-body text-decoration-none')}
            to={getAddURL(service.serviceID)}
            data-test-external-service-card-link={service.kind}
        >
            {children}
        </Link>
    ) : (
        children
    )
}

function getAddURL(id: string): string {
    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}
