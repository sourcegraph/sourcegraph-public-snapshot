import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, ButtonLink } from '@sourcegraph/wildcard'

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

export const ButtonDropdownCta: React.FunctionComponent<ButtonDropdownCtaProps> = ({
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

    const toggleDropdownOpen = useCallback(() => {
        setIsDropdownOpen(isOpen => !isOpen)
        onToggle?.()
    }, [onToggle])

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
        <ButtonDropdown className="menu-nav-item" direction="down" isOpen={isDropdownOpen} toggle={toggleDropdownOpen}>
            <Button
                className={classNames('text-decoration-none', styles.toggle, className)}
                nav={true}
                caret={false}
                variant="secondary"
                outline={true}
                size="sm"
                as={DropdownToggle}
            >
                {button}
            </Button>
            <DropdownMenu right={true} className={styles.container}>
                <div className="d-flex mb-3">
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
            </DropdownMenu>
        </ButtonDropdown>
    )
}
