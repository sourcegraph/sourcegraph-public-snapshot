import EmojiIcon from '@sourcegraph/icons/lib/Emoji'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HUBSPOT_SURVEY_URL = 'https://sourcegraph-2762526.hs-sites.com/user-survey'
const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

export interface SurveyCTAProps {
    className?: string
    score?: number
    onClick?: (score: number) => void
    openSurveyInNewTab?: boolean
}

export class SurveyCTA extends React.PureComponent<SurveyCTAProps> {
    public render(): JSX.Element | null {
        return (
            <div className={this.props.className}>
                {Array(11)
                    .fill(1)
                    .map((_, i) => {
                        const pressed = i === this.props.score
                        if (window.context.hostSurveysLocallyEnabled) {
                            return (
                                <Link
                                    type="button"
                                    key={i}
                                    className={`btn btn-primary toast__rating-btn ${pressed ? 'active' : ''}`}
                                    aria-pressed={pressed || undefined}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={() => this.onClick(i)}
                                    to={`/survey/${i}`}
                                    target={this.props.openSurveyInNewTab ? '_blank' : undefined}
                                >
                                    {i}
                                </Link>
                            )
                        } else {
                            return (
                                <button
                                    type="button"
                                    key={i}
                                    className="btn btn-primary toast__rating-btn"
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={() => this.onClick(i)}
                                >
                                    {i}
                                </button>
                            )
                        }
                    })}
            </div>
        )
    }

    private onClick = (score: number) => {
        eventLogger.log('SurveyButtonClicked', { marketing: { nps_score: score } })
        if (this.props.onClick) {
            this.props.onClick(score)
        }
    }
}

interface State {
    user: GQL.IUser | null
    visible: boolean
}

export class SurveyToast extends React.Component<{}, State> {
    private subscriptions = new Subscription()

    constructor(props: {}) {
        super(props)
        this.state = {
            user: null,
            visible: localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && daysActiveCount === 3,
        }
        if (this.state.visible) {
            eventLogger.log('SurveyReminderViewed', { marketing: { sessionCount: daysActiveCount } })
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<EmojiIcon className="icon-inline" />}
                title="Tell us what you think"
                subtitle="How likely is it that you would recommend Sourcegraph to a friend?"
                cta={
                    <SurveyCTA
                        onClick={
                            window.context.hostSurveysLocallyEnabled ? this.onClickLocalScore : this.onClickRemoteScore
                        }
                        openSurveyInNewTab={true}
                    />
                }
                onDismiss={this.onDismiss}
            />
        )
    }

    private onClickLocalScore = (score: number): void => this.onDismiss()

    private onClickRemoteScore = (score: number) => {
        const url = new URL(HUBSPOT_SURVEY_URL)
        url.searchParams.set('nps_score', score.toString())
        url.searchParams.set('user_is_authenticated', (this.state.user !== null).toString())
        url.searchParams.set('site_id', window.context.siteID)
        if (this.state.user) {
            url.searchParams.set('email', this.state.user.email)
        }
        this.onDismiss()
        window.open(url.href)
    }

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}
