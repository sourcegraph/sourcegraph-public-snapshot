import React, { useEffect } from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import type { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { CodySurveyToast } from '../marketing/toast/CodySurveyToast'
import { eventLogger } from '../tracking/eventLogger'

import { getReturnTo } from './SignInSignUpCommon'
import { withAuthenticatedUser } from './withAuthenticatedUser'

import styles from './PostSignUpPage.module.scss'

interface PostSignUpPageProps {
    authenticatedUser: AuthenticatedUser
}

const PostSignUp: React.FunctionComponent<PostSignUpPageProps> = ({ authenticatedUser }) => {
    const location = useLocation()
    const searchParameters = new URLSearchParams(location.search)
    const isExperimentEnabled = searchParameters.get('experiment_flag')?.toLowerCase() === 'true'

    const containsExperimentFlagParam = searchParameters.has('experiment_flag')
    const shouldRedirect = !containsExperimentFlagParam && authenticatedUser.completedPostSignup
    const [showQualificationSurvey, status] = useFeatureFlag('signup-survey-enabled', false)

    // Redirects if the experiment flag is not provided and if the user has completed the post-signup flow.
    if (shouldRedirect) {
        const returnTo = getReturnTo(location)
        return <Navigate to={returnTo} replace={true} />
    }

    if (status !== 'loaded') {
        return null
    }

    return (
        <div className={styles.pageWrapper}>
            <PageTitle title="Post signup" />
            <Page className={styles.page}>
                <img src="/.assets/img/sourcegraph-mark.svg?v2" alt="Sourcegraph logo" className={styles.logo} />

                <CodySurveyToast
                    telemetryService={eventLogger}
                    authenticatedUser={authenticatedUser}
                    showQualificationSurvey={isExperimentEnabled || showQualificationSurvey}
                />
            </Page>
        </div>
    )
}

export const PostSignUpPage = withAuthenticatedUser(PostSignUp)
