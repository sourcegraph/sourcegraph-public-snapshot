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
        this.state = {
            reason: '',
            betterProduct: '',
            email: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <Form className="survey-form" onSubmit={this.handleSubmit}>
                {this.state.error && <p className="survey-form__error">{this.state.error.message}</p>}
                <label className="survey-form__label">
                    How likely is it that you would recommend Sourcegraph to a friend?
                </label>
                <SurveyCTA className="survey-form__scores" onClick={this.onScoreChange} score={this.props.score} />
                {!this.props.authenticatedUser && (
                    <div className="form-group">
                        <input
                            className="form-control survey-form__input"
                            type="text"
                            placeholder="Email"
                            onChange={this.onEmailFieldChange}
                            value={this.state.email}
                            disabled={this.state.loading}
                        />
                    </div>
                )}
                <div className="form-group">
                    <label className="survey-form__label">
                        What is the most important reason for the score you gave Sourcegraph?
                    </label>
                    <textarea
                        className="form-control survey-form__input"
                        onChange={this.onReasonFieldChange}
                        value={this.state.reason}
                        disabled={this.state.loading}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group">
                    <label className="survey-form__label">What could Sourcegraph do to provide a better product?</label>
                    <textarea
                        className="form-control survey-form__input"
                        onChange={this.onBetterProductFieldChange}
                        value={this.state.betterProduct}
                        disabled={this.state.loading}
                    />
                </div>
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={this.state.loading}>
                        Submit
                    </button>
                </div>
                {this.state.loading && (
                    <div className="survey-form__loader">
                        <LoadingSpinner className="icon-inline" />
                    </div>
                )}
                <div>
                    <small>
                        Your response to this survey will be sent to Sourcegraph, and will be visible to your
                        Sourcegraph site admins.
                    </small>
                </div>
            </Form>
        )
    }

    private onScoreChange = (score: number): void => {
        if (this.props.onScoreChange) {
            this.props.onScoreChange(score)
        }
        this.setState({ error: undefined })
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ email: e.target.value })
    }

    private onReasonFieldChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
        this.setState({ reason: e.target.value })
    }

    private onBetterProductFieldChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
        this.setState({ betterProduct: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        if (this.state.loading) {
            return
        }

        if (this.props.score === undefined) {
            this.setState({ error: new Error('Please select a score') })
            return
        }

        eventLogger.log('SurveySubmitted')
        this.setState({ loading: true })

        this.subscriptions.add(
            submitSurvey({
                score: this.props.score,
                email: this.state.email,
                reason: this.state.reason,
                better: this.state.betterProduct,
            })
                .pipe(
                    catchError(error => {
                        this.setState({ error })
                        return []
                    })
                )
                .subscribe(() => {
                    if (this.props.onSubmit) {
                        this.props.onSubmit()
                    }
                    this.props.history.push('/survey/thanks')
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
        if (this.props.match.params.score === 'thanks') {
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
                    cta={<SurveyForm score={this.intScore(this.props.match.params.score)} {...this.props} />}
                />
            </div>
        )
    }

    private intScore = (score?: string): number | undefined =>
        score ? Math.max(0, Math.min(10, Math.round(+score))) : undefined
}
