import React, { useState } from 'react'

import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'

import { Button, Checkbox, Icon, Input, TextArea } from '@sourcegraph/wildcard'

import { Toast } from './Toast'

import styles from './Step.module.scss'

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

interface UseCaseFormProps {
    handleDismiss: () => void
    handleDone: (props: FormStateType) => void
}

export const UseCaseForm: React.FunctionComponent<UseCaseFormProps> = ({ handleDismiss, handleDone }) => {
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
                            <Button
                                key={id}
                                className={classNames(
                                    styles.checkboxButton,
                                    useCases.includes(id) && styles.checkboxButtonActive
                                )}
                                variant="secondary"
                                outline={true}
                                size="sm"
                            >
                                {useCases.includes(id) && <Icon className={styles.checkboxIcon} as={CheckIcon} />}
                                <Checkbox
                                    onChange={() => handleSelectUseCase(id)}
                                    tabIndex={-1}
                                    id={id}
                                    value={id}
                                    checked={useCases.includes(id)}
                                    label={labelValue}
                                    name={id}
                                />
                            </Button>
                        ))}
                    </div>
                    {useCases.includes('other') && (
                        <Input
                            label="What else are you using sourcegraph to do?"
                            name="other"
                            onChange={event => setOtherUseCase(event.target.value)}
                            placeholder="Find..."
                            value={otherUseCase}
                            className={styles.input}
                        />
                    )}
                    <TextArea
                        className={styles.textarea}
                        name="more"
                        onChange={event => setMoreSharedInfo(event.target.value)}
                        value={moreSharedInfo}
                        label="Anything else you would like to share with us?"
                    />
                </div>
            }
            footer={
                <div className={styles.done}>
                    <Button id="survey-toast-dismiss" variant="primary" size="sm" onClick={() => handleSubmit()}>
                        Done
                    </Button>
                </div>
            }
            onDismiss={() => handleDismiss()}
        />
    )
}
