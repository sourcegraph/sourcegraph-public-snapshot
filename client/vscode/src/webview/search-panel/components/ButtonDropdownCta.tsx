import React, { useCallback, useEffect, useState } from 'react'
import classNames from 'classnames'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, Button } from '@sourcegraph/wildcard'

import styles from './ButtonDropdownCta.module.scss'

// Debt: this is a fork of the web <ButtonDropdownCta>.

export interface ButtonDropdownCtaProps extends TelemetryProps {
    button: JSX.Element
    icon: JSX.Element
    title: string
    copyText: string
    source: string
    viewEventName: string
    returnTo: string
    onToggle?: () => void
    className?: string
    instanceURL: string
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
    instanceURL,
}) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)

    const toggleDropdownOpen = useCallback(() => {
        setIsDropdownOpen(isOpen => !isOpen)
        onToggle?.()
    }, [onToggle])

    const onClick = (): void => {
        telemetryService.log(`VSCE${source}SignUpModalClick`)
    }

    // Whenever dropdown opens, log view event
    useEffect(() => {
        if (isDropdownOpen) {
            telemetryService.log(viewEventName)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isDropdownOpen])

    // Use cloud url instead of instance url
    // This will only display for non private uers
    // Because private users are required use with token = signed up
    const signUpURL = `https://sourcegraph.com/sign-up?src=${source}&returnTo=${encodeURIComponent(
        returnTo
    )}&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up`

    return (
        <ButtonDropdown className="menu-nav-item" direction="down" isOpen={isDropdownOpen} toggle={toggleDropdownOpen}>
            <DropdownToggle
                className={classNames(
                    'btn btn-sm btn-outline-secondary text-decoration-none',
                    className,
                    styles.toggle
                )}
                nav={true}
                caret={false}
            >
                {button}
            </DropdownToggle>
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
                <Button to={signUpURL} onClick={onClick} variant="primary" as={Link}>
                    Sign up for Sourcegraph
                </Button>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
