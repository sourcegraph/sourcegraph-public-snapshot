import { FunctionComponent } from 'react'

import * as H from 'history'

import { Button } from '@sourcegraph/wildcard'

import styles from './PolicyListActions.module.scss'
export interface PolicyListActionsProps {
    history: H.History
}

export const PolicyListActions: FunctionComponent<React.PropsWithChildren<PolicyListActionsProps>> = ({ history }) => (
    <>
        <Button className={styles.btn} variant="primary" onClick={() => history.push('./configuration/new')}>
            Create new policy
        </Button>
    </>
)
