import classNames from 'classnames'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import styles from './PageHeader.module.scss'

type BreadcrumbIcon = React.ComponentType<{ className?: string }>
type BreadcrumbText = React.ReactNode
type Breadcrumb = {
    /** Use a valid path to render this Breadcrumb as a Link */
    to?: string
    icon?: BreadcrumbIcon
    text?: React.ReactNode
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
    headingElement?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
    className?: string
}

export const PageHeader: React.FunctionComponent<Props> = ({
    annotation,
    path,
    byline,
    description,
    actions,
    className,
    headingElement: HeadingX = 'h1',
}) => {
    if (path.length === 0) {
        return null
    }

    return (
        <header className={classNames(styles.pageHeader, className)}>
            <div>
                {annotation && <small className={styles.annotation}>{annotation}</small>}
                <HeadingX className="flex-grow-1 d-block m-0">
                    {path.map(({ to, text, icon: Icon }, index) => (
                        <React.Fragment key={index}>
                            {index !== 0 && <span className="mr-2 text-muted">/</span>}
                            <LinkOrSpan to={to}>
                                {Icon && <Icon className="icon-inline py-1 mr-1" />}
                                {text && <span className="mr-2">{text}</span>}
                            </LinkOrSpan>
                        </React.Fragment>
                    ))}
                </HeadingX>
                {byline && <small className={styles.byline}>{byline}</small>}
                {description && <p className={styles.description}>{description}</p>}
            </div>
            {actions && <div className={styles.actions}>{actions}</div>}
        </header>
    )
}
