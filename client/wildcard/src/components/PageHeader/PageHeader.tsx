import React from 'react'

import classNames from 'classnames'

import { HeadingElement } from '../Typography/Heading/Heading'

import { Breadcrumb, BreadcrumbIcon, BreadcrumbText } from './Breadcrumb'
import { BreadcrumbList } from './BreadcrumbList'

import styles from './PageHeader.module.scss'

type BreadcrumbItem = {
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

type PageHeaderProps = {
    /** Renders small print above the heading */
    annotation?: React.ReactNode
    /** Renders small print below the heading */
    byline?: React.ReactNode
    /** Renders description text below the heading */
    description?: React.ReactNode
    /** Align additional content (e.g. buttons) alongside the heading */
    actions?: React.ReactNode
    /** Heading element to use, defaults to h1 */
    headingElement?: HeadingElement
    className?: string
    children?: React.ReactNode
} & (
    | {
          /** Heading content */
          path: BreadcrumbItem[]
      }
    | {
          /** Heading content */
          children?: React.ReactNode
      }
)

export const PageHeader: React.FunctionComponent<React.PropsWithChildren<PageHeaderProps>> & {
    Breadcrumb: typeof Breadcrumb
    BreadcrumbList: typeof BreadcrumbList
} = props => {
    const { annotation, byline, description, actions, className, children, headingElement } = props
    const path: BreadcrumbItem[] = 'path' in props ? props.path : []

    if (path.length === 0 && !children) {
        return null
    }

    const heading = (
        <BreadcrumbList as={headingElement}>
            {path.map(({ to, text, icon, ariaLabel }, index) => (
                <Breadcrumb key={index} to={to} icon={icon} aria-label={ariaLabel}>
                    {text}
                </Breadcrumb>
            ))}
        </BreadcrumbList>
    )

    return (
        <div className={classNames(styles.container, className)}>
            <div>
                {annotation && <small className={styles.annotation}>{annotation}</small>}
                {children || heading}
                {byline && <small className={styles.byline}>{byline}</small>}
                {description && <p className={styles.description}>{description}</p>}
            </div>
            {actions && <div>{actions}</div>}
        </div>
    )
}

PageHeader.Breadcrumb = Breadcrumb
PageHeader.BreadcrumbList = BreadcrumbList
