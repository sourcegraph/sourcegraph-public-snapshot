import React, { useRef, useState } from 'react'
import TextAreaAutosize from 'react-textarea-autosize'
interface Props extends React.HTMLProps<HTMLTextAreaElement> {}

export const FeedbackPrompt: React.FunctionComponent<Props> = ({ ...textAreaProps }) => (
    // const [content, updateContent] = useState('')
    // const labelParent = useRef<HTMLLabelElement | null>(null)

    // const handleInput: React.ChangeEventHandler<HTMLTextAreaElement> = event => {
    //     console.log(event)
    //     updateContent(event.target.value)
    // }

    <TextAreaAutosize />
)
