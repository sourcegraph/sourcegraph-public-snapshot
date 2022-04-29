import React, { useState } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'

import { Button } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

import radioStyles from './SurveyRatingRadio.module.scss'

interface SurveyRatingRadio {
    ariaLabelledby?: string
    score?: number
    onChange?: (score: number) => void
    openSurveyInNewTab?: boolean
}

export const SurveyRatingRadio: React.FunctionComponent<SurveyRatingRadio> = props => {
    const [selectedIndex, setSelectedIndexIndex] = useState<number | null>(null)

    const handleClick = (score: number): void => {
        eventLogger.log('SurveyButtonClicked', { score }, { score })

        setSelectedIndexIndex(score)
        if (props.onChange) {
            props.onChange(score)
        }
    }

    return (
        <fieldset
            aria-labelledby={props.ariaLabelledby}
            aria-describedby="survey-rating-scale"
            className={radioStyles.scores}
        >
            {range(0, 11).map(score => {
                const selected = score === selectedIndex

                return (
                    <Button
                        key={score}
                        className={classNames(radioStyles.ratingBtn, !selected && radioStyles.ratingBtnDefault)}
                        variant={score === selectedIndex ? 'primary' : 'secondary'}
                        outline={score !== selectedIndex}
                        onClick={() => handleClick(score)}
                    >
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
