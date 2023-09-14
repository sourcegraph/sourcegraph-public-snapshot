import React from 'react'

import classNames from 'classnames'

import { FlexTextArea, H4, Input } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { UseCaseFeedbackModifiers, UseCaseFeedbackState } from '../toast/SurveyUseCaseToast'

import styles from './SurveyUseCaseForm.module.scss'

interface SurveyUseCaseFormProps extends UseCaseFeedbackState, UseCaseFeedbackModifiers {
    formLabelClassName?: string
    className?: string
    authenticatedUser?: AuthenticatedUser | null
}

export const SURVEY_QUESTIONS: Record<'otherUseCase' | 'better' | 'reason', string> = {
    otherUseCase: 'What do you use Sourcegraph for?',
    better: 'How can we make Sourcegraph better?',
    reason: 'What is the most important reason for the score you gave Sourcegraph?',
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({
    otherUseCase,
    onChangeOtherUseCase,
    better,
    onChangeBetter,
    email,
    onChangeEmail,
    formLabelClassName,
    className,
    authenticatedUser,
}) => (
    <div className={classNames('mb-2', className)}>
        <FlexTextArea
            minRows={2}
            maxRows={6}
            containerClassName="mt-3"
            label={
                <H4 as="span" className={classNames('d-flex', styles.title, formLabelClassName)}>
                    {SURVEY_QUESTIONS.otherUseCase}
                </H4>
            }
            onChange={event => onChangeOtherUseCase(event.target.value)}
            value={otherUseCase}
        />
        <FlexTextArea
            minRows={2}
            maxRows={6}
            containerClassName="mt-3"
            label={
                <H4 as="span" className={classNames('d-flex', styles.title, formLabelClassName)}>
                    {SURVEY_QUESTIONS.better}
                </H4>
            }
            onChange={event => onChangeBetter(event.target.value)}
            value={better}
        />
        {!authenticatedUser && (
            <Input
                className="mt-3"
                label={
                    <H4 as="span" className={classNames('d-flex', styles.title, formLabelClassName)}>
                        What is your email?
                    </H4>
                }
                onChange={event => onChangeEmail(event.target.value)}
                value={email}
                type="email"
                autoComplete="email"
                name="email"
            />
        )}
    </div>
)
