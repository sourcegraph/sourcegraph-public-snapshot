import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { WorkspaceEditPreview } from '../threads/detail/inbox/item/WorkspaceEditPreview'

interface Props extends ExtensionsControllerProps {
    /**
     * The code actions to apply.
     */
    codeActions: sourcegraph.CodeAction[]

    className?: string
}

/**
 * A page that shows the impact of applying edits from code actions.
 */
export const ComparePage: React.FunctionComponent<Props> = ({ codeActions, className = '', ...props }) => (
    <div className={`compare-page ${className}`}>
        {codeActions.map(
            (codeAction, i) =>
                codeAction.edit && (
                    <WorkspaceEditPreview
                        key={i}
                        {...props}
                        workspaceEdit={codeAction.edit}
                        className="overflow-auto"
                    />
                )
        )}
    </div>
)
