import { ApolloError, OperationVariables } from '@apollo/client'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import TickIcon from 'mdi-react/TickIcon'
import React, { useCallback, useEffect, useRef, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'

import { Popover, PopoverContent, Position, Button, FlexTextArea, LoadingSpinner, Link } from '../..'
import { useAutoFocus, useLocalStorage } from '../../..'

import { Happy, Sad, VeryHappy, VerySad } from './FeedbackIcons'
import styles from './FeedbackPrompt.module.scss'
import { IconRadioButtons } from './IconRadioButtons'

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

interface ContentProps<TData = any> {
    closePrompt?: () => void
    /** Boolean for displaying the Join Research link */
    productResearchEnabled?: boolean
    onSubmit: (text: string, rating: number) => Promise<OperationVariables | undefined>
    loading: boolean
    data?: TData | null
    error?: ApolloError
}
const LOCAL_STORAGE_KEY_RATING = 'feedbackPromptRating'
const LOCAL_STORAGE_KEY_TEXT = 'feedbackPromptText'

const FeedbackPromptContent: React.FunctionComponent<ContentProps> = ({
    closePrompt,
    productResearchEnabled,
    onSubmit,
    loading,
    data,
    error,
}) => {
    const [rating, setRating] = useLocalStorage<number | undefined>(LOCAL_STORAGE_KEY_RATING, undefined)
    const [text, setText] = useLocalStorage<string>(LOCAL_STORAGE_KEY_TEXT, '')
    const textAreaReference = useRef<HTMLInputElement>(null)
    const handleRateChange = useCallback((value: number) => setRating(value), [setRating])
    const handleTextChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => setText(event.target.value), [
        setText,
    ])

    const handleSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()

            if (!rating) {
                return
            }

            return onSubmit(text, rating)
        },
        [rating, onSubmit, text]
    )

    useEffect(() => {
        if (data?.submitHappinessFeedback) {
            // Reset local storage when successfully submitted
            localStorage.removeItem(LOCAL_STORAGE_KEY_TEXT)
            localStorage.removeItem(LOCAL_STORAGE_KEY_RATING)
        }
    }, [data?.submitHappinessFeedback])

    useAutoFocus({ autoFocus: true, reference: textAreaReference })

    return (
        <>
            <Button className={styles.close} onClick={closePrompt}>
                <CloseIcon className={styles.icon} />
            </Button>
            {data?.submitHappinessFeedback ? (
                <div className={styles.success}>
                    <TickIcon className={styles.successTick} />
                    <h3>We‘ve received your feedback!</h3>
                    <p className="d-inline">
                        Thank you for your help.
                        {productResearchEnabled && (
                            <>
                                {' '}
                                Want to help keep making Sourcegraph better?{' '}
                                <Link to="/user/settings/product-research" onClick={closePrompt}>
                                    Join us for occasional user research
                                </Link>{' '}
                                and share your feedback on our latest features and ideas.
                            </>
                        )}
                    </p>
                </div>
            ) : (
                <Form onSubmit={handleSubmit}>
                    <h3 className="mb-0">What’s on your mind?</h3>

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
                        disabled={loading}
                    />
                    {error && (
                        <ErrorAlert error={error} icon={false} className="mt-3" prefix="Error submitting feedback" />
                    )}
                    <Button
                        disabled={!rating || !text || loading}
                        role="menuitem"
                        type="submit"
                        variant="secondary"
                        data-testid="send-feedback-btn"
                        className={classNames('btn-block', styles.button)}
                    >
                        {loading ? <LoadingSpinner /> : 'Send'}
                    </Button>
                </Form>
            )}
        </>
    )
}

interface Props extends ContentProps {
    open?: boolean
}

export const FeedbackPrompt: React.FunctionComponent<Props> = ({ open, onSubmit, loading, error, data, children }) => {
    const [isOpen, setIsOpen] = useState(() => !!open)
    const handleToggle = useCallback(() => setIsOpen(open => !open), [])
    const forceClose = useCallback(() => setIsOpen(false), [])

    return (
        <div className={styles.feedbackPrompt}>
            <Popover isOpen={isOpen} onOpenChange={handleToggle}>
                {children}
                <PopoverContent position={Position.bottom} className={styles.menu}>
                    <FeedbackPromptContent
                        onSubmit={onSubmit}
                        productResearchEnabled={true}
                        closePrompt={forceClose}
                        data={data}
                        error={error}
                        loading={loading}
                    />
                </PopoverContent>
            </Popover>
        </div>
    )
}
