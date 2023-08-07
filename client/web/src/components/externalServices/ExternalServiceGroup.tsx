import { FC, useState } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Button, Text, Icon, Link } from '@sourcegraph/wildcard'

import { ExternalServiceCard } from './ExternalServiceCard'
import { AddExternalServiceOptions } from './externalServices'

import styles from './ExternalServiceGroup.module.scss'

interface ExternalServiceGroupProps {
    name: string
    services: AddExternalServiceOptions[]
    description: string
    renderServiceIcon: boolean

    icon?: React.ComponentType<{ className?: string }>
    enabled?: boolean
    to?: string

    /**
     * ToIcon is an icon shown on the right-hand side of the card. Default value is right-pointed chevron.
     */
    toIcon?: string | undefined | null
    className?: string
    enabled?: boolean
    badge?: string
    tooltip?: string
    bordered?: boolean
}

export const ExternalServiceGroup: FC<ExternalServiceGroupProps> = ({
    name,
    services,
    description,
    icon,
    renderServiceIcon,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(true)
    const toggleIsOpen = (): void => setIsOpen(prevIsOpen => !prevIsOpen)
    // <Link
    //         className="d-block text-left text-body text-decoration-none"
    //         to={to}
    //         data-test-external-service-card-link={kind}
    //     >
    return (
        <div className={styles.externalServiceGroupContainer}>
            <Button
                className={classNames(styles.externalServiceGroupHeader, {
                    [styles.externalServiceGroupExpandedHeader]: isOpen,
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
                        <li key={index} className={styles.externalServiceGroupNode}>
                            <ExternalServiceCard to={getAddURL(service.id)} {...service} />
                            {/* <Link
                                className={classNames(
                                    styles.externalServiceGroupNodeLink,
                                    'text-left text-body text-decoration-none'
                                )}
                                to={getAddURL(service.id)}
                            >
                                {renderServiceIcon && (
                                    <Icon inline={true} className="mb-0 mr-1" as={service.icon} aria-hidden={true} />
                                )}
                                <Text className={styles.externalServiceGroupNodeDisplayName}>
                                    {service.title}
                                    {'  '}
                                    <span className={styles.externalServiceGroupNodeDescription}>
                                        {service.shortDescription}
                                    </span>
                                </Text>
                            </Link> */}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}

function getAddURL(id: string): string {
    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}
