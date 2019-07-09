import React from 'react'
import { CheckTemplate } from '../../../../../shared/src/api/client/services/checkTemplates'

interface Props {
    /** The HTML root element to use. */
    element?: 'div' | 'li'

    /** The data. */
    checkTemplate: CheckTemplate

    /** A fragment to render at the end. */
    endFragment?: React.ReactFragment

    className?: string
}

/**
 * A check template, with its icon, title, description, etc.
 */
export const CheckTemplateItem: React.FunctionComponent<Props> = ({
    element: Element = 'div',
    checkTemplate: { title, description, icon: Icon, iconColor },
    endFragment,
    className = '',
}) => (
    <Element className={`${className} p-2 d-flex align-items-center justify-content-between`}>
        <div className="media">
            {/* tslint:disable-next-line: jsx-ban-props */}
            {Icon && <Icon className="mr-2 icon-inline h3 mb-0" style={{ color: iconColor }} />}
            <div className="media-body">
                <h4 className="h6 mb-0 text-decoration-none d-flex align-items-center text-body">{title}</h4>
                {description && <p className="mb-0 small text-muted">{description}</p>}
            </div>
        </div>
        {endFragment}
    </Element>
)
