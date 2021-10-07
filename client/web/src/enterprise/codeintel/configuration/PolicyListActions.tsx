import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Button } from '@sourcegraph/wildcard'

import styles from './PolicyListActions.module.scss'
export interface PolicyListActionsProps {
    disabled: boolean
    deleting: boolean
    history: H.History
}

export const PolicyListActions: FunctionComponent<PolicyListActionsProps> = ({ disabled, deleting, history }) => (
    <>
        <Button
            className={styles.btn}
            variant="primary"
            onClick={() => history.push('./configuration/new')}
            disabled={disabled}
        >
            Create new policy
        </Button>

        {deleting && (
            <span className={styles.loading}>
                <LoadingSpinner className="icon-inline" /> Deleting...
            </span>
        )}
    </>
)
