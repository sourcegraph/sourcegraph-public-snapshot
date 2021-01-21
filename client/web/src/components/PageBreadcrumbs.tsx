import React from 'react'
import { Link } from '../../../shared/src/components/Link'

interface Breadcrumb {
    to?: string
    text: string
}

interface Props {
    icon?: React.ComponentType<{ className?: string }>
    path: Breadcrumb[]
}

const Breadcrumb: React.FC<{ to?: string }> = ({ to, children }) => {
    if (to) {
        return <Link to={to}>{children}</Link>
    }
    return <>{children}</>
}

export const PageBreadcrumbs: React.FC<Props> = ({ icon: Icon, path }) => {
    const [first] = path

    return (
        <>
            {Icon && (
                <Breadcrumb to={first.to}>
                    <Icon className="icon-inline" />
                </Breadcrumb>
            )}
            {path.map(({ text, to }) => (
                <React.Fragment key={text}>
                    {' '}
                    / <Breadcrumb to={to}>{text}</Breadcrumb>
                </React.Fragment>
            ))}
        </>
    )
}
