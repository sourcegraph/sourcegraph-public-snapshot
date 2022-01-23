import React from 'react'

import { ContextInvalidatedState } from '../../state'
import { WebviewPageProps } from '../platform/context'

export interface ContextInvalidatedSidebarViewProps extends WebviewPageProps {
    reason: ContextInvalidatedState['context']['reason']
}

export const ContextInvalidatedSidebarView: React.FunctionComponent<ContextInvalidatedSidebarViewProps> = ({
    extensionCoreAPI,
    reason,
}) => (
    <div>
        <h5 className="mt-3 mb-2">Extension Context Invalidated</h5>
        <p>
            Your {reason === 'access-token-change' ? 'access token' : 'Sourcegraph instance URL'} has changed. Please
            reload VS Code to use to Sourcegraph extension.
        </p>
        <button
            type="button"
            className="btn btn-primary font-weight-normal w-100 my-1 border-0"
            onClick={() => extensionCoreAPI.reloadWindow()}
        >
            Reload VS Code
        </button>
    </div>
)
