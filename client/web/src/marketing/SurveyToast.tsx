import classNames from 'classnames'
import { range } from 'lodash'
import EmoticonIcon from 'mdi-react/EmoticonIcon'
import React, { useState } from 'react'
import { useHistory } from 'react-router'

import { AuthenticatedUser } from '../auth'
import { eventLogger } from '../tracking/eventLogger'

import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

interface SurveyCTAProps {
    ariaLabelledby?: string
    className?: string
    score?: number
    onChange?: (score: number) => void
    openSurveyInNewTab?: boolean
    'aria-labelledby'?: string
}

export const SurveyCTA: React.FunctionComponent<SurveyCTAProps> = props => {
    const history = useHistory()
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null)

    const handleFocus = (index: number): void => {
        setFocusedIndex(index)
    }

    const handleBlur = (): void => {
        setFocusedIndex(null)
    }

    const handleChange = (score: number): void => {
        eventLogger.log('SurveyButtonClicked', { score })
        history.push(`/survey/${score}`)

        if (props.onChange) {
            props.onChange(score)
        }
    }

    return (
        <fieldset aria-labelledby={props.ariaLabelledby} className={props.className} onBlur={handleBlur}>
            {range(0, 11).map(score => {
                const pressed = score === props.score
                const focused = score === focusedIndex

                return (
                    <label
                        key={score}
                        className={classNames('btn btn-primary toast__rating-btn', { active: pressed, focus: focused })}
                    >
                        <input
                            type="radio"
                            name="survey-score"
                            value={score}
                            onChange={() => handleChange(score)}
                            onFocus={() => handleFocus(score)}
                            className="toast__rating-radio"
                        />

                        {score}
                    </label>
                )
            })}
        </fieldset>
    )
}

interface Props {
    authenticatedUser: AuthenticatedUser | null
}

interface State {
    visible: boolean
}

export class SurveyToast extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            visible: localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && daysActiveCount % 30 === 3,
        }
        if (this.state.visible) {
            eventLogger.log('SurveyReminderViewed')
        } else if (daysActiveCount % 30 === 0) {
            // Reset toast dismissal 3 days before it will be shown
            localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'false')
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
                cta={<SurveyCTA onChange={this.onChangeScore} openSurveyInNewTab={true} />}
                onDismiss={this.onDismiss}
            />
        )
    }

    private onChangeScore = (): void => this.onDismiss()

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}
