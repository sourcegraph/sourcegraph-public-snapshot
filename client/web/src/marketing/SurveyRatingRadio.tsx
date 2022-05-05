import React, { useState } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'
import { useHistory } from 'react-router'

import { Button } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

import radioStyles from './SurveyRatingRadio.module.scss'

interface SurveyRatingRadio {
    ariaLabelledby?: string
    score?: number
    onChange?: (score: number) => void
    openSurveyInNewTab?: boolean
}

export const SurveyRatingRadio: React.FunctionComponent<React.PropsWithChildren<SurveyRatingRadio>> = props => {
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
        <fieldset
            aria-labelledby={props.ariaLabelledby}
            aria-describedby="survey-rating-scale"
            className={radioStyles.scores}
            onBlur={handleBlur}
        >
            {range(0, 11).map(score => {
                const pressed = score === props.score
                const focused = score === focusedIndex

                return (
                    <Button
                        key={score}
                        className={classNames(radioStyles.ratingBtn, {
                            active: pressed,
                            focus: focused,
                        })}
                        variant="primary"
                        as="label"
                    >
                        <input
                            type="radio"
                            name="survey-score"
                            value={score}
                            onChange={() => handleChange(score)}
                            onFocus={() => handleFocus(score)}
                            className={radioStyles.ratingRadio}
                        />

                        {score}
                    </Button>
                )
            })}
            <div id="survey-rating-scale" className={radioStyles.ratingScale}>
                <small>Not likely at all</small>
                <small>Very likely</small>
            </div>
        </fieldset>
    )
}
