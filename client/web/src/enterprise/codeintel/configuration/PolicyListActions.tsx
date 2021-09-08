import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Button } from '@sourcegraph/wildcard'

export interface PolicyListActionsProps {
    disabled: boolean
    deleting: boolean
    history: H.History
}

export const PolicyListActions: FunctionComponent<PolicyListActionsProps> = ({ disabled, deleting, history }) => (
    <>
        <Button
            className="mt-2"
            variant="primary"
            onClick={() => history.push('./configuration/new')}
            disabled={disabled}
        >
            Create new policy
        </Button>

        {deleting && (
            <span className="ml-2 mt-2">
                <LoadingSpinner className="icon-inline" /> Deleting...
            </span>
        )}
    </>
)
