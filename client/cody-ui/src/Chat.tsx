import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { isDefined } from '@sourcegraph/common'

import { FileLinkProps } from './chat/ContextFiles'
import { ChatInputContext } from './chat/inputContext/ChatInputContext'
import { Transcript } from './chat/Transcript'
import { TranscriptItemClassNames } from './chat/TranscriptItem'

import styles from './Chat.module.css'

interface ChatProps extends ChatClassNames {
    transcript: ChatMessage[]
    messageInProgress: ChatMessage | null
    messageBeingEdited: boolean
    setMessageBeingEdited: (input: boolean) => void
    contextStatus?: ChatContextStatus | null
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
    onSubmit: (text: string) => void
    textAreaComponent: React.FunctionComponent<ChatUITextAreaProps>
    submitButtonComponent: React.FunctionComponent<ChatUISubmitButtonProps>
    suggestionButtonComponent?: React.FunctionComponent<ChatUISuggestionButtonProps>
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    afterTips?: string
    className?: string
    EditButtonContainer?: React.FunctionComponent<EditButtonProps>
    editButtonOnSubmit?: (text: string) => void
    FeedbackButtonsContainer?: React.FunctionComponent<FeedbackButtonsProps>
    feedbackButtonsOnSubmit?: (text: string) => void
    copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit']
    suggestions?: string[]
    setSuggestions?: (suggestions: undefined | []) => void
}

interface ChatClassNames extends TranscriptItemClassNames {
    inputRowClassName?: string
    chatInputContextClassName?: string
    chatInputClassName?: string
}

export interface ChatUITextAreaProps {
    className: string
    rows: number
    autoFocus: boolean
    value: string
    required: boolean
    onInput: React.FormEventHandler<HTMLElement>
    onKeyDown: React.KeyboardEventHandler<HTMLElement>
}

export interface ChatUISubmitButtonProps {
    className: string
    disabled: boolean
    onClick: (event: React.MouseEvent<HTMLButtonElement>) => void
}

export interface ChatUISuggestionButtonProps {
    suggestion: string
    onClick: (event: React.MouseEvent<HTMLButtonElement>) => void
}

export interface EditButtonProps {
    className: string
    disabled?: boolean
    messageBeingEdited: boolean
    setMessageBeingEdited: (input: boolean) => void
}

export interface FeedbackButtonsProps {
    className: string
    disabled?: boolean
    feedbackButtonsOnSubmit: (text: string) => void
}

export interface CopyButtonProps {
    copyButtonOnSubmit: (text: string) => void
}
/**
 * The Cody chat interface, with a transcript of all messages and a message form.
 */
export const Chat: React.FunctionComponent<ChatProps> = ({
    messageInProgress,
    messageBeingEdited,
    setMessageBeingEdited,
    transcript,
    contextStatus,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
    onSubmit,
    textAreaComponent: TextArea,
    submitButtonComponent: SubmitButton,
    suggestionButtonComponent: SuggestionButton,
    fileLinkComponent,
    afterTips,
    className,
    codeBlocksCopyButtonClassName,
    transcriptItemClassName,
    humanTranscriptItemClassName,
    transcriptItemParticipantClassName,
    transcriptActionClassName,
    inputRowClassName,
    chatInputContextClassName,
    chatInputClassName,
    EditButtonContainer,
    editButtonOnSubmit,
    FeedbackButtonsContainer,
    feedbackButtonsOnSubmit,
    copyButtonOnSubmit,
    suggestions,
    setSuggestions,
}) => {
    const [inputRows, setInputRows] = useState(5)
    const [historyIndex, setHistoryIndex] = useState(inputHistory.length)

    const inputHandler = useCallback(
        (inputValue: string): void => {
            const rowsCount = inputValue.match(/\n/g)?.length
            if (rowsCount) {
                setInputRows(rowsCount < 5 ? 5 : rowsCount > 25 ? 25 : rowsCount)
            } else {
                setInputRows(5)
            }
            setFormInput(inputValue)
            if (inputValue !== inputHistory[historyIndex]) {
                setHistoryIndex(inputHistory.length)
            }
        },
        [historyIndex, inputHistory, setFormInput]
    )

    const submitInput = useCallback(
        (input: string): void => {
            if (messageInProgress) {
                return
            }

            onSubmit(input)
            setSuggestions?.(undefined)
            setHistoryIndex(input.length + 1)
            setInputHistory([...inputHistory, input])
        },
        [inputHistory, messageInProgress, onSubmit, setInputHistory, setSuggestions]
    )

    const onChatSubmit = useCallback((): void => {
        // Submit chat only when input is not empty and not in progress
        if (formInput.trim() && !messageInProgress) {
            setInputRows(5)
            setFormInput('')
            submitInput(formInput)
        }
    }, [formInput, messageInProgress, setFormInput, submitInput])

    const onChatKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>): void => {
            // Submit input on Enter press (without shift) and
            // trim the formInput to make sure input value is not empty.
            if (
                event.key === 'Enter' &&
                !event.shiftKey &&
                !event.nativeEvent.isComposing &&
                formInput &&
                formInput.trim()
            ) {
                event.preventDefault()
                event.stopPropagation()
                setMessageBeingEdited(false)
                onChatSubmit()
            }
            // Loop through input history on up arrow press
            if (event.key === 'ArrowUp' && inputHistory.length) {
                if (formInput === inputHistory[historyIndex] || !formInput) {
                    const newIndex = historyIndex - 1 < 0 ? inputHistory.length - 1 : historyIndex - 1
                    setHistoryIndex(newIndex)
                    setFormInput(inputHistory[newIndex])
                }
            }
        },
        [inputHistory, onChatSubmit, formInput, historyIndex, setFormInput, setMessageBeingEdited]
    )

    const transcriptWithWelcome = useMemo<ChatMessage[]>(
        () => [{ speaker: 'assistant', displayText: welcomeText(afterTips) }, ...transcript],
        [afterTips, transcript]
    )

    return (
        <div className={classNames(className, styles.innerContainer)}>
            <Transcript
                transcript={transcriptWithWelcome}
                messageInProgress={messageInProgress}
                messageBeingEdited={messageBeingEdited}
                setMessageBeingEdited={setMessageBeingEdited}
                fileLinkComponent={fileLinkComponent}
                codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                transcriptItemClassName={transcriptItemClassName}
                humanTranscriptItemClassName={humanTranscriptItemClassName}
                transcriptItemParticipantClassName={transcriptItemParticipantClassName}
                transcriptActionClassName={transcriptActionClassName}
                className={styles.transcriptContainer}
                textAreaComponent={TextArea}
                EditButtonContainer={EditButtonContainer}
                editButtonOnSubmit={editButtonOnSubmit}
                FeedbackButtonsContainer={FeedbackButtonsContainer}
                feedbackButtonsOnSubmit={feedbackButtonsOnSubmit}
                copyButtonOnSubmit={copyButtonOnSubmit}
            />

            <form className={classNames(styles.inputRow, inputRowClassName)}>
                {suggestions !== undefined && suggestions.length !== 0 && SuggestionButton ? (
                    <div className={styles.suggestions}>
                        {suggestions.map((suggestion: string) =>
                            suggestion.trim().length > 0 ? (
                                <SuggestionButton
                                    key={suggestion}
                                    suggestion={suggestion}
                                    onClick={() => submitInput(suggestion)}
                                />
                            ) : null
                        )}
                    </div>
                ) : null}
                <div className={styles.textAreaContainer}>
                    <TextArea
                        className={classNames(styles.chatInput, chatInputClassName)}
                        rows={inputRows}
                        value={formInput}
                        autoFocus={true}
                        required={true}
                        onInput={({ target }) => {
                            const { value } = target as HTMLInputElement
                            inputHandler(value)
                        }}
                        onKeyDown={onChatKeyDown}
                    />
                    <SubmitButton
                        className={styles.submitButton}
                        onClick={onChatSubmit}
                        disabled={!!messageInProgress}
                    />
                </div>
                {contextStatus && (
                    <ChatInputContext contextStatus={contextStatus} className={chatInputContextClassName} />
                )}
            </form>
        </div>
    )
}

function welcomeText(afterTips?: string): string {
    return [
        "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.",
        afterTips,
    ]
        .filter(isDefined)
        .join('\n\n')
}
