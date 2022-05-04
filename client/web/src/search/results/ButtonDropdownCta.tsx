import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, Position, Popover, PopoverTrigger, PopoverContent, ButtonLink } from '@sourcegraph/wildcard'

import { CloudSignUpSource } from '../../auth/CloudSignUpPage'

import styles from './ButtonDropdownCta.module.scss'

export interface ButtonDropdownCtaProps extends TelemetryProps {
    button: JSX.Element
    icon: JSX.Element
    title: string
    copyText: string
    source: CloudSignUpSource
    viewEventName: string
    returnTo: string
    onToggle?: () => void
    className?: string
}

export const ButtonDropdownCta: React.FunctionComponent<React.PropsWithChildren<ButtonDropdownCtaProps>> = ({
    button,
    icon,
    title,
    copyText,
    telemetryService,
    source,
    viewEventName,
    returnTo,
    onToggle,
    className,
}) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)

    const toggleDropdownOpen = (isOpen: boolean): void => {
        if (isOpen !== isDropdownOpen) {
            setIsDropdownOpen(isOpen)
            onToggle?.()
        }
    }

    const onClick = (): void => {
        telemetryService.log(`SignUpPLG${source}_1_Search`)
    }

    // Whenever dropdown opens, log view event
    useEffect(() => {
        if (isDropdownOpen) {
            telemetryService.log(viewEventName)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isDropdownOpen])

    return (
        <Popover isOpen={isDropdownOpen} onOpenChange={event => toggleDropdownOpen(event.isOpen)}>
            <PopoverTrigger as={Button} outline={true} variant="secondary" size="sm" className={className}>
                {button}
            </PopoverTrigger>
            <PopoverContent position={Position.bottomEnd} className={classNames(styles.container)}>
                <div className={classNames('d-flex mb-3')}>
                    <div className="d-flex align-items-center mr-3">
                        <div className={styles.icon}>{icon}</div>
                    </div>
                    <div>
                        <div className={styles.title}>
                            <strong>{title}</strong>
                        </div>
                        <div className={classNames('text-muted', styles.copyText)}>{copyText}</div>
                    </div>
                </div>
                <ButtonLink
                    to={buildGetStartedURL('search-dropdown-cta', returnTo)}
                    onClick={onClick}
                    variant="primary"
                >
                    Get started
                </ButtonLink>
            </PopoverContent>
        </Popover>
    )
}
