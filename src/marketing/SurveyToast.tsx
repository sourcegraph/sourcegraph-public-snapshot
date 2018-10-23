import EmoticonIcon from 'mdi-react/EmoticonIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

interface SurveyCTAProps {
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
                        return (
                            <Link
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

interface Props {
    authenticatedUser: GQL.IUser | null
}

interface State {
    visible: boolean
}

export class SurveyToast extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            visible: localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && daysActiveCount === 3,
        }
        if (this.state.visible) {
            eventLogger.log('SurveyReminderViewed', { marketing: { sessionCount: daysActiveCount } })
        }
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<EmoticonIcon className="icon-inline" />}
                title="Tell us what you think"
                subtitle="How likely is it that you would recommend Sourcegraph to a friend?"
                cta={<SurveyCTA onClick={this.onClickScore} openSurveyInNewTab={true} />}
                onDismiss={this.onDismiss}
            />
        )
    }

    private onClickScore = (score: number): void => this.onDismiss()

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}
