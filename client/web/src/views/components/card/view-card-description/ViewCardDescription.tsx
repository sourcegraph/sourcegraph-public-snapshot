import classnames from 'classnames'
import { MdiReactIconComponentType } from 'mdi-react'
import React from 'react'

import styles from './ViewCardDescription.module.scss'

interface ViewCardDescriptionProps {
    title: string
    icon: MdiReactIconComponentType
    className?: string
}

// Since we use react-grid-layout for build draggable insight cards at insight dashboard
// to support text selection within insight card at InsightDescription component we have to
// capture mouse event to prevent all action from react-grid-layout library which will prevent
// default behavior and the text will become unavailable for selection
const stopPropagation: React.MouseEventHandler<HTMLElement> = event => event.stopPropagation()

export const ViewCardDescription: React.FunctionComponent<ViewCardDescriptionProps> = props => {
    const { icon: Icon, title, className = '' } = props

    return (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <small
            title={title}
            className={classnames(styles.viewDescription, 'text-muted', className)}
            onMouseDown={stopPropagation}
        >
            <Icon className="icon-inline" /> {title}
        </small>
    )
}
