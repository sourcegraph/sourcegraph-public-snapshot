import React, { useState } from 'react'

import classNames from 'classnames'

import { FlexTextArea, H4, Input, Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { SurveyUseCase } from '../../graphql-operations'
import { UseCaseFeedbackModifiers, UseCaseFeedbackState } from '../toast/SurveyUseCaseToast'

import { SurveyUseCaseCheckbox } from './SurveyUseCaseCheckbox'

import styles from './SurveyUseCaseForm.module.scss'

export const OPTIONS = [
    {
        id: SurveyUseCase.UNDERSTAND_NEW_CODE,
        labelValue: 'Understand a new codebase',
    },
    {
        id: SurveyUseCase.FIX_SECURITY_VULNERABILITIES,
        labelValue: 'Fix security vulnerabilities',
    },
    {
        id: SurveyUseCase.REUSE_CODE,
        labelValue: 'Reuse code',
    },
    {
        id: SurveyUseCase.RESPOND_TO_INCIDENTS,
        labelValue: 'Respond to incidents',
    },
    {
        id: SurveyUseCase.IMPROVE_CODE_QUALITY,
        labelValue: 'Improve code quality',
    },
] as const

interface SurveyUseCaseFormProps extends UseCaseFeedbackState, UseCaseFeedbackModifiers {
    formLabelClassName?: string
    className?: string
    title: string
    authenticatedUser?: AuthenticatedUser | null
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({
    useCases,
    onChangeUseCases,
    otherUseCase,
    onChangeOtherUseCase,
    better,
    onChangeBetter,
    email,
    onChangeEmail,
    formLabelClassName,
    className,
    title,
    authenticatedUser,
}) => {
    const [showOtherInput, setShowOtherInput] = useState<boolean>(false)

    const handleToggleOtherInput = (shouldShow: boolean): void => {
        if (!shouldShow) {
            // Clear out any entered information in the "What else..." field
            onChangeOtherUseCase('')
        }

        setShowOtherInput(shouldShow)
    }

    const handleUseCaseChange = (value: SurveyUseCase): void => {
        const newUseCases = useCases.includes(value)
            ? useCases.filter(instance => instance !== value)
            : [...useCases, value]

        onChangeUseCases(newUseCases)
    }

    return (
        <div className={classNames('mb-2', className)}>
            <H4 id="usecase-group" className={classNames('d-flex', styles.title, formLabelClassName)}>
                {title}
            </H4>
            <fieldset className={styles.checkWrap} aria-labelledby="usecase-group">
                {OPTIONS.map(({ id, labelValue }) => (
                    <SurveyUseCaseCheckbox
                        label={labelValue}
                        onChange={() => handleUseCaseChange(id)}
                        key={id}
                        id={id}
                        checked={useCases.includes(id)}
                    />
                ))}
                <SurveyUseCaseCheckbox
                    label="Other"
                    onChange={() => handleToggleOtherInput(!showOtherInput)}
                    id="survey_checkbox_other"
                    checked={showOtherInput}
                />
            </fieldset>
            {showOtherInput && (
                <FlexTextArea
                    minRows={2}
                    maxRows={6}
                    containerClassName="mt-3"
                    label={
                        <Text size="small" className={formLabelClassName}>
                            What else do you use Sourcegraph to do?
                        </Text>
                    }
                    onChange={event => onChangeOtherUseCase(event.target.value)}
                    value={otherUseCase}
                />
            )}
            <FlexTextArea
                minRows={2}
                maxRows={6}
                containerClassName="mt-3"
                label={
                    <Text size="small" className={formLabelClassName}>
                        What would make Sourcegraph better?
                    </Text>
                }
                onChange={event => onChangeBetter(event.target.value)}
                value={better}
            />
            {!authenticatedUser && (
                <Input
                    className="mt-3"
                    label={
                        <Text size="small" className={formLabelClassName}>
                            What is your email?
                        </Text>
                    }
                    onChange={event => onChangeEmail(event.target.value)}
                    value={email}
                />
            )}
        </div>
    )
}
