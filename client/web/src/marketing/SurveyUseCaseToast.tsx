import React, { useEffect, useState } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { SurveyUseCase } from '../graphql-operations'

import { SurveyUseCaseForm } from './SurveyUseCaseForm'
import { Toast } from './Toast'

import styles from './SurveyUseCaseToast.module.scss'

interface FormStateType {
    additionalInformation: string
    otherUseCase: string
    useCases: SurveyUseCase[]
}

interface SurveyUseCaseFormToast {
    onDismiss: () => void
    onDone: () => Promise<void>
    onChange: (props: FormStateType) => void
}

export const SurveyUseCaseToast: React.FunctionComponent<SurveyUseCaseFormToast> = ({
    onDismiss,
    onDone,
    onChange,
}) => {
    const [useCases, setUseCases] = useState<SurveyUseCase[]>([])
    const [otherUseCase, setOtherUseCase] = useState<string>('')
    const [additionalInformation, setAdditionalInformation] = useState<string>('')

    const handleSubmit = (): Promise<void> => onDone()

    useEffect(() => {
        onChange({
            useCases,
            otherUseCase,
            additionalInformation,
        })
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [useCases, otherUseCase, additionalInformation])

    return (
        <Toast
            toastBodyClassName={styles.toastBody}
            toastContentClassName="mt-0"
            cta={
                <SurveyUseCaseForm
                    title="You are using sourcegraph to..."
                    onChangeUseCases={value => setUseCases(value)}
                    otherUseCase={otherUseCase}
                    onChangeOtherUseCase={others => setOtherUseCase(others)}
                    additionalInformation={additionalInformation}
                    onChangeMoreShareInfo={moreInfo => setAdditionalInformation(moreInfo)}
                />
            }
            footer={
                <div className="d-flex justify-content-end">
                    <Button variant="primary" size="sm" onClick={handleSubmit}>
                        Done
                    </Button>
                </div>
            }
            onDismiss={onDismiss}
        />
    )
}
