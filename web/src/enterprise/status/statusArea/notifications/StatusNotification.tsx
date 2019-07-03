import H from 'history'
import React, { useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { CreateChangesetFromCodeActionButton } from '../../../tasks/list/item/CreateChangesetFromCodeActionButton'
import { TasksListItemActions } from '../../../tasks/list/item/TasksListItemActions'
import { WorkspaceEditPreview } from '../../../threads/detail/inbox/item/WorkspaceEditPreview'

interface Props extends ExtensionsControllerProps {
    notification: sourcegraph.Notification
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A notification associated with a status.
 */
export const StatusNotification: React.FunctionComponent<Props> = ({ notification, className = '', ...props }) => {
    const [activeCodeAction, setActiveCodeAction] = useState<sourcegraph.CodeAction | undefined>()

    return (
        <div className={`status-notification card ${className}`}>
            <section className="card-body">
                <h4 className="card-title mb-0 font-weight-normal d-flex align-items-center">{notification.message}</h4>
            </section>
            <section className="d-flex">
                {notification.actions && (
                    <TasksListItemActions
                        codeActions={notification.actions}
                        activeCodeAction={activeCodeAction}
                        onCodeActionClick={() => alert('TOOD!(sqs)')}
                        onCodeActionSetActive={setActiveCodeAction}
                        className="pt-2 pb-0"
                        buttonClassName="btn py-0 px-2 text-decoration-none text-left"
                        inactiveButtonClassName="btn-link"
                        activeButtonClassName="border"
                    />
                )}
                <aside
                    className="d-flex flex-column justify-content-between"
                    style={{
                        flex: '2 0 60%',
                        minWidth: '600px',
                        margin: '-0.5rem -1rem -0.5rem 0',
                    }}
                >
                    {activeCodeAction && activeCodeAction.edit ? (
                        <WorkspaceEditPreview
                            key={JSON.stringify(activeCodeAction.edit)}
                            {...props}
                            workspaceEdit={activeCodeAction.edit}
                            className="tasks-list-item__workspace-edit-preview overflow-auto p-2 mb-3"
                        />
                    ) : null}
                </aside>
            </section>
        </div>
    )
}
