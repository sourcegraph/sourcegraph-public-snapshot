import { useState, useCallback } from 'react'

import { gql, useMutation } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Checkbox, Form, H3, Modal, Text, useLocalStorage } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { SubmitCodySurveyResult, SubmitCodySurveyVariables } from '../../graphql-operations'

const SUBMIT_CODY_SURVEY = gql`
    mutation SubmitCodySurvey($isForWork: Boolean!, $isForPersonal: Boolean!) {
        submitCodySurvey(isForWork: $isForWork, isForPersonal: $isForPersonal) {
            alwaysNil
        }
    }
`

const CodySurveyToastInner: React.FC<{ onSubmitEnd: () => void }> = ({ onSubmitEnd }) => {
    const [isCodyForWork, setIsCodyForWork] = useState(false)
    const [isCodyForPersonalStuff, setIsCodyForPersonalStuff] = useState(false)

    const handleCodyForWorkChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForWork(event.target.checked)
    }, [])
    const handleCodyForPersonalStuffChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForPersonalStuff(event.target.checked)
    }, [])

    const [submitCodySurvey, { loading }] = useMutation<SubmitCodySurveyResult, SubmitCodySurveyVariables>(
        SUBMIT_CODY_SURVEY,
        {
            variables: {
                isForWork: isCodyForWork,
                isForPersonal: isCodyForPersonalStuff,
            },
        }
    )

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>) => {
            event.preventDefault()
            // eslint-disable-next-line no-console
            submitCodySurvey().catch(console.error).finally(onSubmitEnd)
        },
        [onSubmitEnd, submitCodySurvey]
    )

    return (
        <Modal position="center" aria-label="Welcome message">
            <H3 className="mb-4">Quick question...</H3>
            <Text className="mb-3">How will you be using Cody, our AI assistant?</Text>
            <Form onSubmit={handleSubmit}>
                <Checkbox
                    id="cody-for-work"
                    label="for work"
                    wrapperClassName="mb-2"
                    checked={isCodyForWork}
                    disabled={loading}
                    onChange={handleCodyForWorkChange}
                />
                <Checkbox
                    id="cody-for-personal"
                    label="for personal stuff"
                    wrapperClassName="mb-2"
                    checked={isCodyForPersonalStuff}
                    disabled={loading}
                    onChange={handleCodyForPersonalStuffChange}
                />
                <div className="d-flex justify-content-end">
                    <LoaderButton variant="primary" type="submit" loading={loading} label="Get started" />
                </div>
            </Form>
        </Modal>
    )
}

export const useCodySurveyToast = (): {
    show: boolean
    dismiss: () => void
    setShouldShowCodySurvey: (show: boolean) => void
} => {
    // we specifically use local storage as we want consistent value between when user is logged out and logged in / signed up
    // eslint-disable-next-line no-restricted-syntax
    const [shouldShowCodySurvey, setShouldShowCodySurvey] = useLocalStorage('cody.survey.show', false)
    const [hasSubmitted, setHasSubmitted] = useTemporarySetting('cody.survey.submitted', false)
    const dismiss = useCallback(() => setHasSubmitted(true), [setHasSubmitted])

    return {
        // we calculate "show" value based whether this a new signup and whether they already have submitted survey
        show: !hasSubmitted && shouldShowCodySurvey,
        dismiss,
        setShouldShowCodySurvey,
    }
}

export const CodySurveyToast: React.FC = () => {
    const { show, dismiss } = useCodySurveyToast()
    if (!show) {
        return null
    }

    return <CodySurveyToastInner onSubmitEnd={dismiss} />
}
