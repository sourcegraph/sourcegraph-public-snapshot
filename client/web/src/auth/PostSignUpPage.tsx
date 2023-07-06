import React from 'react'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { CodySurveyToast } from '../marketing/toast/CodySurveyToast'
import { eventLogger } from '../tracking/eventLogger'

import styles from './PostSignUpPage.module.scss'

interface PostSignUpPageProps {
    authenticatedUser: AuthenticatedUser | null
}

export const PostSignUpPage: React.FunctionComponent<PostSignUpPageProps> = ({ authenticatedUser }) => (
    <div className={styles.pageWrapper}>
        <Page className={styles.page}>
            <img
                src="https://sourcegraph.com/.assets/img/sourcegraph-mark.svg?v2"
                alt="Sourcegraph logo"
                className={styles.logo}
            />
            {authenticatedUser && (
                <CodySurveyToast telemetryService={eventLogger} authenticatedUser={authenticatedUser} />
            )}
        </Page>
    </div>
)
