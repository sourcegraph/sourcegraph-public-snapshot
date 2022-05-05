import React, { useState } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { SurveyUseCaseForm } from './SurveyUseCaseForm'
import { Toast } from './Toast'

import styles from './SurveyUseCaseToast.module.scss'

interface FormStateType {
    moreSharedInfo: string
    otherUseCase: string
    useCases: string[]
}

interface SurveyUseCaseFormToast {
    onDismiss: () => void
    handleDone: (props: FormStateType) => void
}

export const SurveyUseCaseToast: React.FunctionComponent<SurveyUseCaseFormToast> = ({ onDismiss, handleDone }) => {
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
                    moreSharedInfo={moreSharedInfo}
                    onChangeMoreShareInfo={moreInfo => setMoreSharedInfo(moreInfo)}
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
