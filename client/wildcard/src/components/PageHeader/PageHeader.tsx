import React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { HeadingElement, Heading } from '../Typography'

import styles from './PageHeader.module.scss'

type BreadcrumbIcon = React.ComponentType<{ className?: string }>
type BreadcrumbText = React.ReactNode
type Breadcrumb = {
    /** Use a valid path to render this Breadcrumb as a Link */
    to?: string
    icon?: BreadcrumbIcon
    text?: React.ReactNode
    ariaLabel?: string
} & (
    | {
          icon: BreadcrumbIcon
      }
    | {
          text: BreadcrumbText
      }
)

interface Props {
    /** Renders small print above the heading */
    annotation?: React.ReactNode
    /** Heading content */
    path: Breadcrumb[]
    /** Renders small print below the heading */
    byline?: React.ReactNode
    /** Renders description text below the heading */
    description?: React.ReactNode
    /** Align additional content (e.g. buttons) alongside the heading */
    actions?: React.ReactNode
    /** Heading element to use, defaults to h1 */
    headingElement?: HeadingElement
    /** Set heading element style */
    headingStyleAs?: HeadingElement
    className?: string
}

export const PageHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    annotation,
    path,
    byline,
    description,
    actions,
    className,
    headingElement = 'h1',
    headingStyleAs,
}) => {
    if (path.length === 0) {
        return null
    }

    return (
        <div className={classNames(styles.container, className)}>
            <div>
                {annotation && <small className={styles.annotation}>{annotation}</small>}
                <Heading as={headingElement} styleAs={headingStyleAs ?? headingElement} className={styles.heading}>
                    {path.map(({ to, text, icon: Icon, ariaLabel }, index) => (
                        <React.Fragment key={index}>
                            {index !== 0 && <span className={styles.divider}>/</span>}
                            <LinkOrSpan to={to} className={styles.path} aria-label={ariaLabel}>
                                {Icon && <Icon className={styles.pathIcon} />}
                                {text && <span className={styles.pathText}>{text}</span>}
                            </LinkOrSpan>
                        </React.Fragment>
                    ))}
                </Heading>
                {byline && <small className={styles.byline}>{byline}</small>}
                {description && <p className={styles.description}>{description}</p>}
            </div>
            {actions && <div>{actions}</div>}
        </div>
    )
}
