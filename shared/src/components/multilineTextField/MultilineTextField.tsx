import React from 'react'
import TextareaAutosize, { TextareaAutosizeProps } from 'react-textarea-autosize'
import { Key } from 'ts-key-enum'

interface Props extends Pick<TextareaAutosizeProps, Exclude<keyof TextareaAutosizeProps, 'ref'>> {}

/**
 * A text field that starts as a single line and grows to multiple lines as needed fit its contents,
 * like the message field in Slack. Pressing Enter "submits" it; pressing Shift+Enter
 * inserts a line break.
 */
export const MultilineTextField: React.FunctionComponent<Props> = props => {
    const onKeyPress: React.KeyboardEventHandler<HTMLTextAreaElement> = e => {
        if (e.key === Key.Enter && !e.shiftKey) {
            e.preventDefault()
            if (e.currentTarget.form) {
                e.currentTarget.form.dispatchEvent(new Event('submit'))
            }
        }
    }
    return <TextareaAutosize {...props} onKeyPress={onKeyPress} />
}
