import React, { useState } from 'react'

import classNames from 'classnames'

import { Button, FlexTextArea, TextArea } from '@sourcegraph/wildcard'

import { SurveyUseCaseCheckbox } from './SurveyUseCaseCheckbox'
import { Toast } from './Toast'

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

interface FormStateType {
    moreSharedInfo: string
    otherUseCase: string
    useCases: string[]
}

interface SurveyUseCaseFormProps {
    handleDismiss: () => void
    handleDone: (props: FormStateType) => void
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({ handleDismiss, handleDone }) => {
    const [useCases, setUseCases] = useState<string[]>([])
    const [otherUseCase, setOtherUseCase] = useState<string>('')
    const [moreSharedInfo, setMoreSharedInfo] = useState<string>('')

    const handleSubmit = (): void => {
        handleDone({
            useCases,
            otherUseCase,
            moreSharedInfo,
        })
    }

    const handleSelectUseCase = (value: string): void => {
        if (useCases.includes(value)) {
            setUseCases(current => current.filter(instance => instance !== value))
            return
        }
        setUseCases(current => [...current, value])
    }

    return (
        <Toast
            className={styles.toast}
            subtitle={<span className={styles.toastSubtitle}>You are using sourcegraph to...</span>}
            cta={
                <div className="mb-2">
                    <div className={styles.checkWrap}>
                        {OPTIONS.map(({ id, labelValue }) => (
                            <SurveyUseCaseCheckbox
                                onChange={() => handleSelectUseCase(id)}
                                key={id}
                                label={labelValue}
                            />
                        ))}
                    </div>
                    {useCases.includes('other') && (
                        <FlexTextArea
                            containerClassName={classNames('mt-3', styles.textarea)}
                            label="What else are you using sourcegraph to do?"
                            name="other"
                            placeholder="Find..."
                            onChange={event => setOtherUseCase(event.target.value)}
                            value={otherUseCase}
                        />
                    )}
                    <TextArea
                        className={classNames('mt-3', styles.textarea)}
                        name="more"
                        onChange={event => setMoreSharedInfo(event.target.value)}
                        value={moreSharedInfo}
                        label="Anything else you would like to share with us?"
                    />
                </div>
            }
            footer={
                <div className="d-flex justify-content-end">
                    <Button variant="primary" size="sm" onClick={() => handleSubmit()}>
                        Done
                    </Button>
                </div>
            }
            onDismiss={() => handleDismiss()}
        />
    )
}
