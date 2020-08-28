import * as React from 'react'

interface Props {
    title: string
    state: 'loading' | 'content' | 'empty'
    loadingDisplay: JSX.Element
    contentDisplay: JSX.Element
    emptyDisplay: JSX.Element
    actionButtons?: JSX.Element
    className?: string
}

export const PanelContainer: React.FunctionComponent<Props> = ({
    title,
    state,
    loadingDisplay,
    contentDisplay,
    emptyDisplay,
    actionButtons,
    className,
}) => (
    <div className={`${className || ''} panel-container`}>
        <div className="panel-container__header">
            <h3 className="panel-container__header-text">{title}</h3>
            {actionButtons}
        </div>

        {state === 'loading' && loadingDisplay}
        {state === 'content' && contentDisplay}
        {state === 'empty' && emptyDisplay}
    </div>
)
