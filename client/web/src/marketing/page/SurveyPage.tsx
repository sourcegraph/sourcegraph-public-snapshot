import React, { useEffect } from 'react'

import { useParams, useLocation } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { FeedbackText } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { TweetFeedback } from '../components/TweetFeedback'

import { SurveyForm } from './SurveyForm'

import styles from './SurveyPage.module.scss'

interface SurveyPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
    /**
     * For Storybook only
     */
    forceScore?: string
}

const getScoreFromString = (score?: string): number | undefined =>
    score ? Math.max(0, Math.min(10, Math.round(+score))) : undefined

export const SurveyPage: React.FunctionComponent<React.PropsWithChildren<SurveyPageProps>> = props => {
    const location = useLocation()
    const matchParameters = useParams<{ score?: string }>()
    const score = props.forceScore || matchParameters.score

    useEffect(() => {
        EVENT_LOGGER.logViewEvent('Survey')
        props.telemetryRecorder.recordEvent('surveyNPS.page', 'view')
    }, [props.telemetryRecorder])

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
            <HeroPage
                title="Almost there..."
                cta={
                    <SurveyForm
                        score={getScoreFromString(score)}
                        authenticatedUser={props.authenticatedUser}
                        telemetryRecorder={props.telemetryRecorder}
                    />
                }
            />
        </div>
    )
}
