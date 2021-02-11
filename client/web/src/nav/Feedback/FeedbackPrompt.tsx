import MessageDrawIcon from 'mdi-react/MessageDrawIcon'
import TickIcon from 'mdi-react/TickIcon'
import React, { useCallback, useState, useEffect } from 'react'
import TextAreaAutosize from 'react-textarea-autosize'
import { Alert, ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { gql } from '../../../../shared/src/graphql/graphql'
import { LoaderButton } from '../../components/LoaderButton'
import { SubmitSurveyResult, SubmitSurveyVariables } from '../../graphql-operations'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import { useMutation } from '../../hooks/useMutation'
import { IconRadioButtons } from '../IconRadioButtons'
import { Happy, Sad, VeryHappy, VerySad } from './FeedbackIcons'

const SUBMIT_FEEDBACK_QUERY = gql`
    mutation SubmitSurvey($input: SurveySubmissionInput!) {
        submitSurvey(input: $input) {
            alwaysNil
        }
    }
`

const FeedbackPromptContent: React.FunctionComponent = () => {
    const [rating, setRating] = useLocalStorage<number | undefined>('feedbackPromptRating', undefined)
    const [text, setText] = useLocalStorage<string>('feedbackPromptText', '')
    const handleRateChange = useCallback((value: number) => setRating(value), [setRating])
    const handleTextChange = useCallback(
        (event: React.ChangeEvent<HTMLTextAreaElement>) => setText(event.target.value),
        [setText]
    )
    const [submitFeedback, { loading, data, error }] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(
        SUBMIT_FEEDBACK_QUERY
    )

    const handleSubmit = useCallback((): void => {
        if (rating) {
            return submitFeedback({ input: { score: rating, reason: text } })
        }
    }, [rating, submitFeedback, text])

    useEffect(() => {
        if (data) {
            // Reset local storage for future submissions
            setText('')
            setRating(undefined)
        }
    }, [data, setRating, setText])

    return (
        <DropdownMenu right={true} className="p-3 feedback-prompt__menu align-middle">
            {data && (
                <div className="feedback-prompt__success">
                    <TickIcon className="feedback-prompt__success--tick" />
                    <h3>We've received your feedback!</h3>
                    <p className="d-inline">
                        Thank you for your help.{' '}
                        {window.context.productResearchPageEnabled && (
                            <>
                                Want to help keep making Sourcegraph better?{' '}
                                <a href="/settings/product-research">Join us for occasional user research.</a> and share
                                your feedback on our latest features and ideas.
                            </>
                        )}
                    </p>
                </div>
            )}
            {!data && (
                <>
                    <h3>What's on your mind?</h3>
                    <TextAreaAutosize
                        role="menuitem"
                        onChange={handleTextChange}
                        value={text}
                        minRows={3}
                        maxRows={6}
                        placeholder="What's going well? What could be better?"
                        className="form-control feedback-prompt__textarea"
                    />
                    <IconRadioButtons
                        name="emoji-selector"
                        icons={[
                            {
                                name: 'Very sad',
                                value: 1,
                                icon: VerySad,
                            },
                            {
                                name: 'Sad',
                                value: 2,
                                icon: Sad,
                            },
                            {
                                name: 'Happy',
                                value: 3,
                                icon: Happy,
                            },
                            {
                                name: 'Very Happy',
                                value: 4,
                                icon: VeryHappy,
                            },
                        ]}
                        selected={rating}
                        onChange={handleRateChange}
                        disabled={loading}
                    />

                    {error && (
                        <Alert className="mt-3 feedback-prompt__alert" color="danger">
                            Something went wrong while sending your feedback. Please try again.
                        </Alert>
                    )}
                    <LoaderButton
                        role="menuitem"
                        className="w-100 btn btn-block btn-secondary mt-3"
                        loading={loading}
                        label="Send"
                        onClick={handleSubmit}
                        disabled={!rating}
                    />
                </>
            )}
        </DropdownMenu>
    )
}

export const FeedbackPrompt: React.FunctionComponent<{}> = () => {
    const [isOpen, setIsOpen] = useState(false)
    const handleToggle = useCallback(() => setIsOpen(open => !open), [])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={handleToggle} className="border feedback-prompt">
            <DropdownToggle
                caret={false}
                className="btn btn-link text-decoration-none"
                nav={true}
                aria-label="Feedback"
            >
                <MessageDrawIcon className="d-lg-none icon-inline" />
                <span className="d-none d-lg-block">Feedback</span>
            </DropdownToggle>
            {isOpen && <FeedbackPromptContent />}
        </ButtonDropdown>
    )
}
