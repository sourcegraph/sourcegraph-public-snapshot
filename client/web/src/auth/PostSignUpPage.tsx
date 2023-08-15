import React from 'react'

import { Navigate } from 'react-router-dom'

import type { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { CodySurveyToast } from '../marketing/toast/CodySurveyToast'
import { eventLogger } from '../tracking/eventLogger'

import { withAuthenticatedUser } from './withAuthenticatedUser'

import styles from './PostSignUpPage.module.scss'

interface PostSignUpPageProps {
    authenticatedUser: AuthenticatedUser
}

const PostSignUp: React.FunctionComponent<PostSignUpPageProps> = ({ authenticatedUser }) => {
    // Redirects to /search page, if user has already completed post signup flow
    if (authenticatedUser.completedPostSignup) {
        return <Navigate to="/search" replace={true} />
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
