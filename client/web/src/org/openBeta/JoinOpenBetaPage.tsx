import classNames from 'classnames'
import React from 'react'
import { RouteComponentProps } from 'react-router'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import styles from './JoinOpenBeta.module.scss'

interface Props extends RouteComponentProps {
    authenticatedUser: AuthenticatedUser
}

export const JoinOpenBetaPage: React.FunctionComponent<Props> = () => (
    <li data-test-membersheader="memberslist-header">
        <div className="d-flex align-items-center justify-content-between">
            <div
                className={classNames(
                    'd-flex align-items-center justify-content-start flex-1 member-details',
                    styles.test
                )}
            >
                test title
            </div>
            <div className={styles.test}>Role</div>
            <div className={styles.test} />
        </div>
    </li>
)
