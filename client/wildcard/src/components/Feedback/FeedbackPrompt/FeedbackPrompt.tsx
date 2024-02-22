import React, { type ReactNode, useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiCheck } from '@mdi/js'
import classNames from 'classnames'

import { Popover, PopoverContent, Position, Button, FlexTextArea, LoadingSpinner, Link, H3, Text } from '../..'
import { useAutoFocus, useLocalStorage } from '../../../hooks'
import { ErrorAlert } from '../../ErrorAlert'
import { Form } from '../../Form'
import { Icon } from '../../Icon'
import { Modal } from '../../Modal'

import styles from './FeedbackPrompt.module.scss'

interface FeedbackPromptSubmitResponse {
    isHappinessFeedback?: boolean
    errorMessage?: string
}

export type FeedbackPromptSubmitEventHandler = (text: string) => Promise<FeedbackPromptSubmitResponse>

interface FeedbackPromptAuthenticatedUserProps {
    authenticatedUser: {
        username: string
        email: string
    } | null
}

interface FeedbackPromptContentProps extends FeedbackPromptAuthenticatedUserProps {
    onClose?: () => void
    /** Boolean for displaying the Join Research link */
    productResearchEnabled?: boolean
    onSubmit: FeedbackPromptSubmitEventHandler
    initialValue?: string
}
const LOCAL_STORAGE_KEY_TEXT = 'feedbackPromptText'

const FeedbackPromptContent: React.FunctionComponent<React.PropsWithChildren<FeedbackPromptContentProps>> = ({
    onClose,
    productResearchEnabled,
    onSubmit,
    authenticatedUser,
    initialValue,
}) => {
    const [text, setText] = useLocalStorage<string>(LOCAL_STORAGE_KEY_TEXT, initialValue || '')
    const textAreaReference = useRef<HTMLInputElement>(null)
    const handleTextChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => setText(event.target.value),
        [setText]
    )

    const [submitting, setSubmitting] = useState(false)
    const [submitResponse, setSubmitResponse] = useState<FeedbackPromptSubmitResponse>()

    const handleSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            setSubmitting(true)

            const { isHappinessFeedback, errorMessage } = await onSubmit(text)

            setSubmitResponse({
                isHappinessFeedback,
                errorMessage,
            })

            setSubmitting(false)
        },
        [onSubmit, text]
    )

    useEffect(() => {
        if (submitResponse?.isHappinessFeedback) {
            // Reset local storage when successfully submitted
            localStorage.removeItem(LOCAL_STORAGE_KEY_TEXT)
        }
    }, [submitResponse])

    useAutoFocus({ autoFocus: true, reference: textAreaReference })

    return (
        <>
            <Button className={styles.close} onClick={onClose}>
                <Icon inline={false} svgPath={mdiClose} className={styles.icon} aria-label="Close" />
            </Button>
            {submitResponse?.isHappinessFeedback ? (
                <div className={styles.success}>
                    <Icon inline={false} svgPath={mdiCheck} className={styles.successTick} aria-label="Success" />
                    <H3>We‘ve received your feedback!</H3>
                    <Text className="d-inline">
                        Thank you.
                        {productResearchEnabled && authenticatedUser && (
                            <>
                                {' '}
                                Want to help keep making Sourcegraph better?{' '}
                                <Link to="/user/settings/product-research" onClick={onClose}>
                                    Join us for occasional user research
                                </Link>{' '}
                                and share your feedback on our latest features and ideas.
                            </>
                        )}
                    </Text>
                </div>
            ) : (
                <Form onSubmit={handleSubmit}>
                    <H3 className="mb-3" id="feedback-prompt-question">
                        Send feedback to Sourcegraph
                    </H3>

                    {authenticatedUser && (
                        <Text className={styles.from} size="small">
                            From: {authenticatedUser.username} ({authenticatedUser.email})
                        </Text>
                    )}
                    <FlexTextArea
                        aria-labelledby="feedback-prompt-question"
                        onChange={handleTextChange}
                        value={text}
                        minRows={3}
                        maxRows={6}
                        className={classNames(styles.textarea, authenticatedUser ? styles.textareaWithFrom : '')}
                        autoFocus={true}
                        ref={textAreaReference}
                    />
                    {!authenticatedUser && (
                        <Text className="text-muted" size="small">
                            You're not signed in. Please tell us how to contact you if you want a reply.
                        </Text>
                    )}

                    {submitResponse?.errorMessage && (
                        <ErrorAlert
                            error={submitResponse?.errorMessage}
                            className="mt-3"
                            prefix="Error submitting feedback"
                        />
                    )}
                    <Text className="d-flex align-items-center justify-content-between mt-2">
                        <span>
                            By submitting your feedback, you agree to the{' '}
                            <Link to="https://sourcegraph.com/terms/privacy">Sourcegraph Privacy Policy</Link>.
                        </span>
                    </Text>
                    <Button
                        disabled={!text || submitting}
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
    children?: React.FunctionComponent<React.PropsWithChildren<FeedbackPromptTriggerProps>> | ReactNode
    initialValue?: string
}

export const FeedbackPrompt: React.FunctionComponent<FeedbackPromptProps> = ({
    openByDefault = false,
    onSubmit,
    children,
    onClose,
    position = Position.bottomEnd,
    modal = false,
    modalLabelId = 'sourcegraph-feedback-modal',
    productResearchEnabled,
    authenticatedUser,
    initialValue,
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
            authenticatedUser={authenticatedUser}
            initialValue={initialValue}
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
