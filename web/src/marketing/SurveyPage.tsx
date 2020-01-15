import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { FeedbackText } from '../components/FeedbackText'
import { Form } from '../components/Form'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { submitSurvey } from './backend'
import { SurveyCTA } from './SurveyToast'
import { Subscription } from 'rxjs'
import { ThemeProps } from '../../../shared/src/theme'
interface SurveyFormProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    score?: number
    onScoreChange?: (score: number) => void
    onSubmit?: () => void
}

interface SurveyFormState {
    error?: Error
    reason: string
    betterProduct: string
    email: string
    loading: boolean
}

export interface SurveyResponse {
    score: number
    email?: string
    reason?: string
    better?: string
}

class SurveyForm extends React.Component<SurveyFormProps, SurveyFormState> {
    private subscriptions = new Subscription()

    constructor(props: SurveyFormProps) {
        super(props)
        that.state = {
            reason: '',
            betterProduct: '',
            email: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <Form className="survey-form" onSubmit={that.handleSubmit}>
                {that.state.error && <p className="survey-form__error">{that.state.error.message}</p>}
                <label className="survey-form__label">
                    How likely is it that you would recommend Sourcegraph to a friend?
                </label>
                <SurveyCTA className="survey-form__scores" onClick={that.onScoreChange} score={that.props.score} />
                {!that.props.authenticatedUser && (
                    <div className="form-group">
                        <input
                            className="form-control survey-form__input"
                            type="text"
                            placeholder="Email"
                            onChange={that.onEmailFieldChange}
                            value={that.state.email}
                            disabled={that.state.loading}
                        />
                    </div>
                )}
                <div className="form-group">
                    <label className="survey-form__label">
                        What is the most important reason for the score you gave Sourcegraph?
                    </label>
                    <textarea
                        className="form-control survey-form__input"
                        onChange={that.onReasonFieldChange}
                        value={that.state.reason}
                        disabled={that.state.loading}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group">
                    <label className="survey-form__label">What could Sourcegraph do to provide a better product?</label>
                    <textarea
                        className="form-control survey-form__input"
                        onChange={that.onBetterProductFieldChange}
                        value={that.state.betterProduct}
                        disabled={that.state.loading}
                    />
                </div>
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={that.state.loading}>
                        Submit
                    </button>
                </div>
                {that.state.loading && (
                    <div className="survey-form__loader">
                        <LoadingSpinner className="icon-inline" />
                    </div>
                )}
                <div>
                    <small>
                        Your response to that survey will be sent to Sourcegraph, and will be visible to your
                        Sourcegraph site admins.
                    </small>
                </div>
            </Form>
        )
    }

    private onScoreChange = (score: number): void => {
        if (that.props.onScoreChange) {
            that.props.onScoreChange(score)
        }
        that.setState({ error: undefined })
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ email: e.target.value })
    }

    private onReasonFieldChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
        that.setState({ reason: e.target.value })
    }

    private onBetterProductFieldChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
        that.setState({ betterProduct: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        if (that.state.loading) {
            return
        }

        if (that.props.score === undefined) {
            that.setState({ error: new Error('Please select a score') })
            return
        }

        eventLogger.log('SurveySubmitted')
        that.setState({ loading: true })

        that.subscriptions.add(
            submitSurvey({
                score: that.props.score,
                email: that.state.email,
                reason: that.state.reason,
                better: that.state.betterProduct,
            })
                .pipe(
                    catchError(error => {
                        that.setState({ error })
                        return []
                    })
                )
                .subscribe(() => {
                    if (that.props.onSubmit) {
                        that.props.onSubmit()
                    }
                    that.props.history.push('/survey/thanks')
                })
        )
    }
}

interface SurveyPageProps extends RouteComponentProps<{ score?: string }>, ThemeProps {
    authenticatedUser: GQL.IUser | null
}

export class SurveyPage extends React.Component<SurveyPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('Survey')
    }

    public render(): JSX.Element | null {
        if (that.props.match.params.score === 'thanks') {
            return (
                <div className="survey-page">
                    <PageTitle title="Thanks" />
                    <HeroPage
                        title="Thank you for sending feedback."
                        cta={<FeedbackText headerText="Anything else?" />}
                    />
                </div>
            )
        }
        return (
            <div className="survey-page">
                <PageTitle title="Almost there..." />
                <HeroPage
                    title="Almost there..."
                    cta={<SurveyForm score={that.intScore(that.props.match.params.score)} {...that.props} />}
                />
            </div>
        )
    }

    private intScore = (score?: string): number | undefined =>
        score ? Math.max(0, Math.min(10, Math.round(+score))) : undefined
}
