import classNames from 'classnames'
import { range } from 'lodash'
import React, { useState } from 'react'
import { useHistory } from 'react-router'

import { eventLogger } from '../tracking/eventLogger'

import toastStyles from './Toast.module.scss'

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
        eventLogger.log('SurveyButtonClicked', { score }, { score })
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
                        className={classNames('btn btn-primary', toastStyles.ratingBtn, {
                            active: pressed,
                            focus: focused,
                        })}
                    >
                        <input
                            type="radio"
                            name="survey-score"
                            value={score}
                            onChange={() => handleChange(score)}
                            onFocus={() => handleFocus(score)}
                            className={toastStyles.ratingRadio}
                        />

                        {score}
                    </label>
                )
            })}
        </fieldset>
    )
}
