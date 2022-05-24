import React, { ReactNode, useCallback, useEffect, useRef, useState } from 'react'

import CloseIcon from 'mdi-react/CloseIcon'
import TickIcon from 'mdi-react/TickIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'

import { Popover, PopoverContent, Position, Button, FlexTextArea, LoadingSpinner, Link, Typography } from '../..'
import { useAutoFocus, useLocalStorage } from '../../..'
import { Modal } from '../../Modal'

import { Happy, Sad, VeryHappy, VerySad } from './FeedbackIcons'
import { IconRadioButtons } from './IconRadioButtons'

import styles from './FeedbackPrompt.module.scss'

export const HAPPINESS_FEEDBACK_OPTIONS = [
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
]

interface FeedbackPromptSubmitResponse {
    isHappinessFeedback?: boolean
    errorMessage?: string
}

export type FeedbackPromptSubmitEventHandler = (text: string, rating: number) => Promise<FeedbackPromptSubmitResponse>

interface FeedbackPromptContentProps {
    onClose?: () => void
    /** Boolean for displaying the Join Research link */
    productResearchEnabled?: boolean
    onSubmit: FeedbackPromptSubmitEventHandler
}
const LOCAL_STORAGE_KEY_RATING = 'feedbackPromptRating'
const LOCAL_STORAGE_KEY_TEXT = 'feedbackPromptText'

const FeedbackPromptContent: React.FunctionComponent<React.PropsWithChildren<FeedbackPromptContentProps>> = ({
    onClose,
    productResearchEnabled,
    onSubmit,
}) => {
    const [rating, setRating] = useLocalStorage<number | undefined>(LOCAL_STORAGE_KEY_RATING, undefined)
    const [text, setText] = useLocalStorage<string>(LOCAL_STORAGE_KEY_TEXT, '')
    const textAreaReference = useRef<HTMLInputElement>(null)
    const handleRateChange = useCallback((value: number) => setRating(value), [setRating])
    const handleTextChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => setText(event.target.value), [
        setText,
    ])

    const [submitting, setSubmitting] = useState(false)
    const [submitResponse, setSubmitResponse] = useState<FeedbackPromptSubmitResponse>()

    const handleSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            if (!rating) {
                return
            }

            setSubmitting(true)

            const { isHappinessFeedback, errorMessage } = await onSubmit(text, rating)

            setSubmitResponse({
                isHappinessFeedback,
                errorMessage,
            })

            setSubmitting(false)
        },
        [rating, onSubmit, text]
    )

    useEffect(() => {
        if (submitResponse?.isHappinessFeedback) {
            // Reset local storage when successfully submitted
            localStorage.removeItem(LOCAL_STORAGE_KEY_TEXT)
            localStorage.removeItem(LOCAL_STORAGE_KEY_RATING)
        }
    }, [submitResponse])

    useAutoFocus({ autoFocus: true, reference: textAreaReference })

    return (
        <>
            <Button className={styles.close} onClick={onClose}>
                <CloseIcon className={styles.icon} />
            </Button>
            {submitResponse?.isHappinessFeedback ? (
                <div className={styles.success}>
                    <TickIcon className={styles.successTick} />
                    <Typography.H3>We‘ve received your feedback!</Typography.H3>
                    <p className="d-inline">
                        Thank you for your help.
                        {productResearchEnabled && (
                            <>
                                {' '}
                                Want to help keep making Sourcegraph better?{' '}
                                <Link to="/user/settings/product-research" onClick={onClose}>
                                    Join us for occasional user research
                                </Link>{' '}
                                and share your feedback on our latest features and ideas.
                            </>
                        )}
                    </p>
                </div>
            ) : (
                <Form onSubmit={handleSubmit}>
                    <Typography.H3 className="mb-3">What’s on your mind?</Typography.H3>

                    <FlexTextArea
                        onChange={handleTextChange}
                        value={text}
                        minRows={3}
                        maxRows={6}
                        placeholder="What’s going well? What could be better?"
                        className={styles.textarea}
                        autoFocus={true}
                        ref={textAreaReference}
                    />

                    <IconRadioButtons
                        name="emoji-selector"
                        icons={HAPPINESS_FEEDBACK_OPTIONS}
                        selected={rating}
                        onChange={handleRateChange}
                        disabled={submitting}
                        className="mt-3"
                    />
                    {submitResponse?.errorMessage && (
                        <ErrorAlert
                            error={submitResponse?.errorMessage}
                            icon={false}
                            className="mt-3"
                            prefix="Error submitting feedback"
                        />
                    )}
                    <Button
                        disabled={!rating || !text || submitting}
                        role="menuitem"
                        type="submit"
                        display="block"
                        variant="secondary"
                        className={styles.button}
                    >
                        {submitting ? <LoadingSpinner /> : 'Send'}
                    </Button>
                </Form>
            )}
        </>
    )
}

interface FeedbackPromptTriggerProps {
    isOpen?: boolean
    onClick?: React.MouseEventHandler<HTMLElement>
}

interface FeedbackPromptProps extends FeedbackPromptContentProps {
    /**
     * Determines if the prompt is opened by default
     *
     * @default false
     */
    openByDefault?: boolean
    position?: Position
    modal?: boolean
    modalLabelId?: string
    children: React.FunctionComponent<React.PropsWithChildren<FeedbackPromptTriggerProps>> | ReactNode
}

export const FeedbackPrompt: React.FunctionComponent<React.PropsWithChildren<FeedbackPromptProps>> = ({
    openByDefault = false,
    onSubmit,
    children,
    onClose,
    position = Position.bottomEnd,
    modal = false,
    modalLabelId = 'sourcegraph-feedback-modal',
    productResearchEnabled,
}) => {
    const [isOpen, setIsOpen] = useState(() => !!openByDefault)
    const ChildrenComponent = typeof children === 'function' && children
    const handleClosePrompt = useCallback(() => {
        setIsOpen(false)
        onClose?.()
    }, [onClose])

    const toggleIsOpen = useCallback(() => {
        setIsOpen(isOpen => !isOpen)
    }, [])

    const triggerElement = ChildrenComponent ? <ChildrenComponent isOpen={isOpen} onClick={toggleIsOpen} /> : children

    const contentElement = (
        <FeedbackPromptContent
            onSubmit={onSubmit}
            productResearchEnabled={productResearchEnabled}
            onClose={handleClosePrompt}
        />
    )

    if (modal) {
        return (
            <>
                {triggerElement}
                {isOpen && (
                    <Modal onDismiss={handleClosePrompt} aria-labelledby={modalLabelId}>
                        {contentElement}
                    </Modal>
                )}
            </>
        )
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
            {triggerElement}
            <PopoverContent position={position} className={styles.menu}>
                {contentElement}
            </PopoverContent>
        </Popover>
    )
}
