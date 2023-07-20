import React from 'react'

import { Navigate } from 'react-router-dom'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { CodySurveyToast } from '../marketing/toast/CodySurveyToast'
import { PageRoutes } from '../routes.constants'
import { eventLogger } from '../tracking/eventLogger'

import { withAuthenticatedUser } from './withAuthenticatedUser'

import styles from './PostSignUpPage.module.scss'

interface PostSignUpPageProps {
    authenticatedUser: AuthenticatedUser
}

const PostSignUp: React.FunctionComponent<PostSignUpPageProps> = ({ authenticatedUser }) => {
    // redirects user to /get-cody page if they have aleady completed the post signup flow.
    if (authenticatedUser.completedPostSignup) {
        return <Navigate to={PageRoutes.GetCody} replace={true} />
    }

    return (
        <div className={styles.pageWrapper}>
            <PageTitle title="Post signup" />
            <Page className={styles.page}>
                <img src="/.assets/img/sourcegraph-mark.svg?v2" alt="Sourcegraph logo" className={styles.logo} />

                <CodySurveyToast telemetryService={eventLogger} authenticatedUser={authenticatedUser} />
            </Page>
        </div>
    )
}

export const PostSignUpPage = withAuthenticatedUser(PostSignUp)
