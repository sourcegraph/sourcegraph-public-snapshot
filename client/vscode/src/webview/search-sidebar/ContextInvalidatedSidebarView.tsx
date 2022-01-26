import React from 'react'

import { WebviewPageProps } from '../platform/context'

export interface ContextInvalidatedSidebarViewProps extends WebviewPageProps {}

export const ContextInvalidatedSidebarView: React.FunctionComponent<ContextInvalidatedSidebarViewProps> = ({
    extensionCoreAPI,
}) => (
    <div>
        <h5 className="mt-3 mb-2">Your Sourcegraph instance URL has changed.</h5>
        <p>Please reload VS Code to use to Sourcegraph extension.</p>
        <button
            type="button"
            className="btn btn-primary font-weight-normal w-100 my-1 border-0"
            onClick={() => extensionCoreAPI.reloadWindow()}
        >
            Reload VS Code
        </button>
    </div>
)
