import React from 'react'
import classNames from 'classnames'

interface Props {
    annotation?: React.ReactNode
    title: React.ReactNode
    byline?: React.ReactNode
    icon?: React.ComponentType<{ className?: string }>
    actions?: React.ReactNode
    className?: string
}

const Muted: React.FC<{ className?: string }> = ({ children, className }) => (
    <div className={classNames('text-muted mb-3', className)}>{children}</div>
)

export const PageHeader: React.FunctionComponent<Props> = ({
    annotation,
    title,
    byline,
    icon: Icon,
    actions,
    className,
}) => (
    <div
        className={classNames(
            'page-header d-flex flex-column flex-md-row flex-wrap justify-content-between align-items-lg-center mb-3 mb-md-0',
            className
        )}
    >
        <div>
            {annotation && <Muted>{annotation}</Muted>}
            <h1 className="flex-grow-1 d-block">
                {Icon && <Icon className="icon-inline" />} {title}
            </h1>
            {byline && <Muted>{byline}</Muted>}
        </div>
        {actions && <div>{actions}</div>}
    </div>
)
