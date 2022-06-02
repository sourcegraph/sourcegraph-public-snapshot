import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { FlexTextArea, H4, TextArea } from '@sourcegraph/wildcard'

import { SurveyUseCase } from '../graphql-operations'

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
]

interface SurveyUseCaseFormProps {
    onChangeUseCases: (useCases: SurveyUseCase[]) => void
    onChangeOtherUseCase: (others: string) => void
    onChangeMoreShareInfo: (moreInfo: string) => void
    formLabelClassName?: string
    additionalInformation: string
    otherUseCase: string
    className?: string
    title: string
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({
    onChangeMoreShareInfo,
    onChangeOtherUseCase,
    onChangeUseCases,
    formLabelClassName,
    additionalInformation,
    otherUseCase,
    className,
    title,
}) => {
    const [useCases, setUseCases] = useState<SurveyUseCase[]>([])
    const [showOtherInput, setShowOtherInput] = useState<boolean>(false)

    const handleSelectUseCase = (value: SurveyUseCase): void => {
        if (useCases.includes(value)) {
            setUseCases(current => current.filter(instance => instance !== value))
            return
        }
        setUseCases(current => [...current, value])
    }

    useEffect(() => {
        onChangeUseCases(useCases)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [useCases])

    return (
        <div className={classNames('mb-2', className)}>
            <H4 id="usecase-group" className={classNames('d-flex', styles.title, formLabelClassName)}>
                {title}
            </H4>
            <div className={styles.checkWrap}>
                {OPTIONS.map(({ id, labelValue }) => (
                    <SurveyUseCaseCheckbox
                        onChange={() => handleSelectUseCase(id)}
                        key={id}
                        id={id}
                        checked={useCases.includes(id)}
                        label={labelValue}
                        aria-labelledby="usecase-group"
                    />
                ))}
                <SurveyUseCaseCheckbox
                    onChange={() => setShowOtherInput(!showOtherInput)}
                    key="other"
                    id="other"
                    checked={showOtherInput}
                    label="other"
                    aria-labelledby="usecase-group"
                />
            </div>
            {showOtherInput && (
                <FlexTextArea
                    containerClassName="mt-3"
                    label={
                        <span className={classNames(styles.textareaLabel, formLabelClassName)}>
                            What else are you using sourcegraph to do?
                        </span>
                    }
                    name="other"
                    placeholder="Find..."
                    onChange={event => onChangeOtherUseCase(event.target.value)}
                    value={otherUseCase}
                />
            )}
            <TextArea
                className="mt-3"
                size="small"
                name="more"
                onChange={event => onChangeMoreShareInfo(event.target.value)}
                value={additionalInformation}
                label={
                    <span className={classNames(styles.textareaLabel, formLabelClassName)}>
                        Anything else you would like to share with us?
                    </span>
                }
            />
        </div>
    )
}
