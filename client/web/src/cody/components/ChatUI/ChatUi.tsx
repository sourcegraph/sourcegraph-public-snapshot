import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiSend, mdiArrowDown, mdiPencil } from '@mdi/js'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { Chat, ChatUISubmitButtonProps, ChatUITextAreaProps, EditButtonProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'
import { Button, Icon, TextArea } from '@sourcegraph/wildcard'

import { useChatStoreState } from '../../stores/chat'

import styles from './ChatUi.module.scss'

export const SCROLL_THRESHOLD = 100

export const ChatUI = (): JSX.Element => {
    const { submitMessage, editMessage, messageInProgress, transcript, getChatContext, transcriptId } =
        useChatStoreState()

    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [messageBeingEdited, setMessageBeingEdited] = useState<boolean>(false)

    return (
        <Chat
            key={transcriptId}
            messageInProgress={messageInProgress}
            messageBeingEdited={messageBeingEdited}
            setMessageBeingEdited={setMessageBeingEdited}
            transcript={transcript}
            formInput={formInput}
            setFormInput={setFormInput}
            inputHistory={inputHistory}
            setInputHistory={setInputHistory}
            onSubmit={submitMessage}
            contextStatus={getChatContext()}
            submitButtonComponent={SubmitButton}
            fileLinkComponent={FileLink}
            className={styles.container}
            afterTips={CODY_TERMS_MARKDOWN}
            transcriptItemClassName={styles.transcriptItem}
            humanTranscriptItemClassName={styles.humanTranscriptItem}
            transcriptItemParticipantClassName="text-muted"
            inputRowClassName={styles.inputRow}
            chatInputClassName={styles.chatInput}
            EditButtonContainer={EditButton}
            editButtonOnSubmit={editMessage}
            textAreaComponent={AutoResizableTextArea}
            codeBlocksCopyButtonClassName={styles.codeBlocksCopyButton}
            transcriptActionClassName={styles.transcriptAction}
        />
    )
}

export const ScrollDownButton = ({ onClick }: { onClick: () => void }): JSX.Element => (
    <div className={styles.scrollButtonWrapper}>
        <Button className={styles.scrollButton} onClick={onClick}>
            <Icon aria-label="Scroll down" svgPath={mdiArrowDown} />
        </Button>
    </div>
)

export const EditButton: React.FunctionComponent<EditButtonProps> = ({
    className,
    messageBeingEdited,
    setMessageBeingEdited,
}) => (
    <div className={className}>
        <button
            className={classNames(className, styles.editButton)}
            type="button"
            onClick={() => setMessageBeingEdited(!messageBeingEdited)}
        >
            {messageBeingEdited ? (
                <Icon aria-label="Close" svgPath={mdiClose} />
            ) : (
                <Icon aria-label="Edit" svgPath={mdiPencil} />
            )}
        </button>
    </div>
)

export const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={classNames(className, styles.submitButton)} type="submit" disabled={disabled} onClick={onClick}>
        <Icon aria-label="Submit" svgPath={mdiSend} />
    </button>
)

export const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>

interface AutoResizableTextAreaProps extends ChatUITextAreaProps {}

export const AutoResizableTextArea: React.FC<AutoResizableTextAreaProps> = ({
    value,
    onInput,
    onKeyDown,
    className,
}) => {
    const textAreaRef = useRef<HTMLTextAreaElement>(null)
    const { width = 0 } = useResizeObserver({ ref: textAreaRef })

    const adjustTextAreaHeight = useCallback((): void => {
        if (textAreaRef.current) {
            textAreaRef.current.style.height = '0px'
            const scrollHeight = textAreaRef.current.scrollHeight
            textAreaRef.current.style.height = `${scrollHeight}px`

            // Hide scroll if the textArea isn't overflowing.
            textAreaRef.current.style.overflowY = scrollHeight < 200 ? 'hidden' : 'auto'
        }
    }, [])

    const handleChange = (): void => {
        adjustTextAreaHeight()
    }

    useEffect(() => {
        adjustTextAreaHeight()
    }, [adjustTextAreaHeight, value, width])

    return (
        <TextArea
            ref={textAreaRef}
            className={className}
            value={value}
            onChange={handleChange}
            rows={1}
            autoFocus={false}
            required={true}
            onKeyDown={onKeyDown}
            onInput={onInput}
        />
    )
}
