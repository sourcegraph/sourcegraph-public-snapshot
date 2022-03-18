import React from 'react'
import classNames from 'classnames'

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
}) => (
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
        <a
            className={classNames('btn', styles.btn)}
            href="https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up"
            onClick={onClickAction}
        >
            <span className={styles.text}>{buttonText}</span>
        </a>
    </div>
)
