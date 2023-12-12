import React from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import type { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { CodySurveyToast } from '../marketing/toast/CodySurveyToast'
import { eventLogger, telemetryRecorder } from '../tracking/eventLogger'

import { getReturnTo } from './SignInSignUpCommon'
import { withAuthenticatedUser } from './withAuthenticatedUser'

import styles from './PostSignUpPage.module.scss'

interface PostSignUpPageProps {
    authenticatedUser: AuthenticatedUser
}

const PostSignUp: React.FunctionComponent<PostSignUpPageProps> = ({ authenticatedUser }) => {
    const location = useLocation()

    // Redirect if user has already completed post signup flow
    if (authenticatedUser.completedPostSignup) {
        const returnTo = getReturnTo(location)
        return <Navigate to={returnTo} replace={true} />
    }

    return (
        <div className={styles.pageWrapper}>
            <PageTitle title="Post signup" />
            <Page className={styles.page}>
                <img src="/.assets/img/sourcegraph-mark.svg?v2" alt="Sourcegraph logo" className={styles.logo} />

                <CodySurveyToast
                    telemetryService={eventLogger}
                    telemetryRecorder={telemetryRecorder}
                    authenticatedUser={authenticatedUser}
                />
            </Page>
        </div>
    )
}

export const PostSignUpPage = withAuthenticatedUser(PostSignUp)
