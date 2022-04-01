import React, { useState } from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Icon, Link } from '@sourcegraph/wildcard'

import styles from './SearchCta.module.scss'

interface SearchPageCtaProps {
    icon: JSX.Element
    ctaTitle: string
    ctaDescription: string
    buttonText: string
    onClickAction?: () => void
}

export const SearchPageCta: React.FunctionComponent<SearchPageCtaProps> = ({
    icon,
    ctaTitle,
    ctaDescription,
    buttonText,
    onClickAction,
}) => {
    const [show, setShow] = useState(true)

    if (!show) {
        return null
    }

    return (
        <div className={classNames(styles.container)}>
            <div className="mr-md-3 ml-3">
                <div className="w-50">{icon}</div>
            </div>
            <div className={classNames('flex-1 my-md-0 my-2', styles.contentContainer)}>
                <div className={classNames('mb-1', styles.title)}>
                    <strong>{ctaTitle}</strong>
                </div>
                <div className={classNames('text-muted', styles.description)}>{ctaDescription}</div>
            </div>
            <Button
                type="button"
                variant="primary"
                className={classNames('btn', styles.btn)}
                as={Link}
                to="https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up"
                onClick={onClickAction}
            >
                <span className={styles.text}>{buttonText}</span>
            </Button>
            <Button
                onClick={() => setShow(false)}
                variant="icon"
                className={classNames(styles.dismiss)}
                title="Close panel"
                data-tooltip="Close panel"
                data-placement="left"
            >
                <Icon as={CloseIcon} />
            </Button>
        </div>
    )
}
