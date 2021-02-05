import React, { useRef, useState } from 'react'

interface Props extends React.HTMLProps<HTMLTextAreaElement> {}

export const GrowingTextArea: React.FunctionComponent<Props> = ({ ...textAreaProps }) => {
    const [content, updateContent] = useState('')
    const labelParent = useRef<HTMLLabelElement | null>(null)

    const handleInput: React.ChangeEventHandler<HTMLTextAreaElement> = event => {
        console.log(event)
        updateContent(event.target.value)
    }

    return (
        <label className="growing-textarea stacked" ref={labelParent} data-value={content}>
            <textarea onInput={handleInput} {...textAreaProps} />
        </label>
    )
}
