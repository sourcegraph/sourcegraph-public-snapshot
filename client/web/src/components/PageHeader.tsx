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
    annotation?: React.ReactNode
    path: Breadcrumb[]
    byline?: React.ReactNode
    actions?: React.ReactNode
    className?: string
}

export const PageHeader: React.FunctionComponent<Props> = ({ annotation, path, byline, actions, className }) => {
    if (path.length === 0) {
        return null
    }

    return (
        <header
            className={classNames(
                'page-header d-flex flex-column flex-md-row flex-wrap justify-content-between align-items-lg-center',
                className
            )}
        >
            <div>
                {annotation && <div className="text-muted page-header__annotation">{annotation}</div>}
                <h1 className="flex-grow-1 d-block m-0">
                    {path.map(({ to, text, icon: Icon }, index) => (
                        <React.Fragment key={index}>
                            {index !== 0 && <span className="mr-2">/</span>}
                            <LinkOrSpan to={to}>
                                {Icon && <Icon className="icon-inline page-header__icon mr-2" />}
                                {text && <span className="mr-2">{text}</span>}
                            </LinkOrSpan>
                        </React.Fragment>
                    ))}
                </h1>
                {byline && <div className="text-muted page-header__byline">{byline}</div>}
            </div>
            {actions && <div className="mt-3 mt-md-0">{actions}</div>}
        </header>
    )
}
