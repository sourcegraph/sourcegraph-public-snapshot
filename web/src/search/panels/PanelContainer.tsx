import * as React from 'react'

interface Props {
    title: string
    state: 'loading' | 'populated' | 'empty'
    // the content displayed when state is 'loading'
    loadingContent: JSX.Element
    // the content displayed when state is 'populated'
    populatedContent: JSX.Element
    // the content displayed when state is 'empty'
    emptyContent: JSX.Element
    actionButtons?: JSX.Element
    className?: string
}

export const PanelContainer: React.FunctionComponent<Props> = ({
    title,
    state,
    loadingContent: loadingDisplay,
    populatedContent: contentDisplay,
    emptyContent: emptyDisplay,
    actionButtons,
    className,
}) => (
    <div className={`${className || ''} panel-container`}>
        <div className="panel-container__header d-flex">
            <h3 className="panel-container__header-text">{title}</h3>
            {actionButtons}
        </div>

        {state === 'loading' && loadingDisplay}
        {state === 'populated' && contentDisplay}
        {state === 'empty' && emptyDisplay}
    </div>
)
