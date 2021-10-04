import React, { useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { FeedbackText } from '../components/FeedbackText'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

import { SurveyForm } from './SurveyForm'
import styles from './SurveyPage.module.scss'
import { TweetFeedback } from './TweetFeedback'

interface SurveyPageProps extends RouteComponentProps<{ score?: string }>, ThemeProps {
    authenticatedUser: AuthenticatedUser | null
}

// TODO: Better naming
const intScore = (score?: string): number | undefined =>
    score ? Math.max(0, Math.min(10, Math.round(+score))) : undefined

export const SurveyPage: React.FunctionComponent<SurveyPageProps> = props => {
    const {
        location,
        match: {
            params: { score },
        },
    } = props

    useEffect(() => {
        eventLogger.logViewEvent('Survey')
    }, [])

    if (score === 'thanks') {
        return (
            <div className={styles.surveyPage}>
                <PageTitle title="Thanks" />
                <HeroPage
                    title="Thanks for the feedback!"
                    body={<TweetFeedback score={location.state.score} feedback={location.state.feedback} />}
                    cta={<FeedbackText headerText="Anything else?" />}
                />
            </div>
        )
    }

    return (
        <div className={styles.surveyPage}>
            <PageTitle title="Almost there..." />
            <HeroPage title="Almost there..." cta={<SurveyForm score={intScore(score)} {...props} />} />
        </div>
    )
}
