import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import TickIcon from 'mdi-react/TickIcon'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { gql, useMutation } from '@sourcegraph/shared/src/graphql/graphql'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { Button, FlexTextArea, LoadingSpinner, useAutoFocus } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../components/alerts'
import { SubmitHappinessFeedbackResult, SubmitHappinessFeedbackVariables } from '../../graphql-operations'
import { useRoutesMatch } from '../../hooks'
import { LayoutRouteProps } from '../../routes'
import { IconRadioButtons } from '../IconRadioButtons'

import { Happy, Sad, VeryHappy, VerySad } from './FeedbackIcons'
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

export const SUBMIT_HAPPINESS_FEEDBACK_QUERY = gql`
    mutation SubmitHappinessFeedback($input: HappinessFeedbackSubmissionInput!) {
        submitHappinessFeedback(input: $input) {
            alwaysNil
        }
    }
`

interface ContentProps {
    closePrompt: () => void
    routeMatch?: string
    /** Text to be prepended to user input on submission. */
    textPrefix?: string
    /** Boolean for displaying the Join Research link */
    productResearchEnabled?: boolean
}

const LOCAL_STORAGE_KEY_RATING = 'feedbackPromptRating'
const LOCAL_STORAGE_KEY_TEXT = 'feedbackPromptText'

export const FeedbackPromptContent: React.FunctionComponent<ContentProps> = ({
    closePrompt,
    routeMatch,
    productResearchEnabled,
    textPrefix = '',
}) => {
    const [rating, setRating] = useLocalStorage<number | undefined>(LOCAL_STORAGE_KEY_RATING, undefined)
    const [text, setText] = useLocalStorage<string>(LOCAL_STORAGE_KEY_TEXT, '')
    const textAreaReference = useRef<HTMLInputElement>(null)
    const handleRateChange = useCallback((value: number) => setRating(value), [setRating])
    const handleTextChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => setText(event.target.value), [
        setText,
    ])

    const [submitFeedback, { loading, data, error }] = useMutation<
        SubmitHappinessFeedbackResult,
        SubmitHappinessFeedbackVariables
    >(SUBMIT_HAPPINESS_FEEDBACK_QUERY)

    const handleSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()

            if (!rating) {
                return
            }

            return submitFeedback({
                variables: {
                    input: { score: rating, feedback: `${textPrefix}${text}`, currentPath: routeMatch },
                },
            })
        },
        [rating, submitFeedback, text, routeMatch, textPrefix]
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
                        className={classNames('btn-block', styles.button)}
                    >
                        {loading ? <LoadingSpinner /> : 'Send'}
                    </Button>
                </Form>
            )}
        </>
    )
}

interface Props {
    open?: boolean
    routes: readonly LayoutRouteProps<{}>[]
}

export const FeedbackPrompt: React.FunctionComponent<Props> = ({ open, routes }) => {
    const [isOpen, setIsOpen] = useState(() => !!open)
    const handleToggle = useCallback(() => setIsOpen(open => !open), [])
    const forceClose = useCallback(() => setIsOpen(false), [])
    const match = useRoutesMatch(routes)

    return (
        <ButtonDropdown
            a11y={false}
            isOpen={isOpen}
            toggle={handleToggle}
            className={styles.feedbackPrompt}
            group={false}
        >
            <DropdownToggle
                tag="button"
                caret={false}
                className={classNames('btn btn-sm btn-outline-secondary text-decoration-none', styles.toggle)}
                aria-label="Feedback"
            >
                <span>Feedback</span>
            </DropdownToggle>
            <DropdownMenu right={true} className={styles.menu}>
                <FeedbackPromptContent productResearchEnabled={true} closePrompt={forceClose} routeMatch={match} />
            </DropdownMenu>
        </ButtonDropdown>
    )
}
