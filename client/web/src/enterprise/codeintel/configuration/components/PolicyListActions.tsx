import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { Button } from '@sourcegraph/wildcard'

import styles from './PolicyListActions.module.scss'
export interface PolicyListActionsProps {
    history: H.History
}

export const PolicyListActions: FunctionComponent<PolicyListActionsProps> = ({ history }) => (
    <>
        <Button className={styles.btn} variant="primary" onClick={() => history.push('./configuration/new')}>
            Create new policy
        </Button>
    </>
)
