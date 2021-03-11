import React from 'react'
import classNames from 'classnames'
import { LinkOrSpan } from '../../../shared/src/components/LinkOrSpan'

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
    actions,
    className,
    headingElement: HeadingX = 'h1',
}) => {
    if (path.length === 0) {
        return null
    }

    return (
        <header
            className={classNames(
                'd-flex flex-column flex-md-row flex-wrap justify-content-between align-items-lg-center',
                className
            )}
        >
            <div>
                {annotation && <small className="text-muted d-block mb-2">{annotation}</small>}
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
                {byline && <small className="text-muted d-block mt-1">{byline}</small>}
            </div>
            {actions && <div className="mt-3 mt-md-0">{actions}</div>}
        </header>
    )
}
