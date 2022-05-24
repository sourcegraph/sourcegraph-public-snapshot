import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { FlexTextArea, Typography, TextArea } from '@sourcegraph/wildcard'

import { SurveyUseCaseCheckbox } from './SurveyUseCaseCheckbox'

import styles from './SurveyUseCaseForm.module.scss'

export const OPTIONS = [
    {
        id: 'understandCodebase',
        labelValue: 'Understand a new codebase',
    },
    {
        id: 'fixSecurityVulnerability',
        labelValue: 'Fix security vulnerabilities',
    },
    {
        id: 'reuseCode',
        labelValue: 'Reuse code',
    },
    {
        id: 'respondToIncidents',
        labelValue: 'Respond to incidents',
    },
    {
        id: 'ImproveCodeQuality',
        labelValue: 'Improve code quality',
    },
    {
        id: 'other',
        labelValue: 'Other',
    },
]

interface SurveyUseCaseFormProps {
    onChangeUseCases: (useCases: string[]) => void
    onChangeOtherUseCase: (others: string) => void
    onChangeMoreShareInfo: (moreInfo: string) => void
    formLabelClassName?: string
    moreSharedInfo: string
    otherUseCase: string
    className?: string
    title: string
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({
    onChangeMoreShareInfo,
    onChangeOtherUseCase,
    onChangeUseCases,
    formLabelClassName,
    moreSharedInfo,
    otherUseCase,
    className,
    title,
}) => {
    const [useCases, setUseCases] = useState<string[]>([])

    const handleSelectUseCase = (value: string): void => {
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
            <Typography.H4 id="usecase-group" className={classNames('d-flex', styles.title, formLabelClassName)}>
                {title}
            </Typography.H4>
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
            </div>
            {useCases.includes('other') && (
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
                value={moreSharedInfo}
                label={
                    <span className={classNames(styles.textareaLabel, formLabelClassName)}>
                        Anything else you would like to share with us?
                    </span>
                }
            />
        </div>
    )
}
