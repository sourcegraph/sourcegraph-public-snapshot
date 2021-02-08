import MessageDrawIcon from 'mdi-react/MessageDrawIcon'
import React, { useCallback, useState } from 'react'
import { Button, ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import TextAreaAutosize from 'react-textarea-autosize'

interface Props {}

export const FeedbackPrompt: React.FC<Props> = () => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className="border feedback-prompt">
            <DropdownToggle caret={false} className="btn btn-link" nav={true} aria-label="Feedback">
                <MessageDrawIcon className="d-lg-none icon-inline" />
                <span className="d-none d-lg-block">Feedback</span>
            </DropdownToggle>

            <DropdownMenu right={true} className="p-3 feedback-prompt__menu">
                <h3 className="align-middle">What's on your mind?</h3>
                <TextAreaAutosize
                    minRows={3}
                    maxRows={6}
                    placeholder="What's going well? What could be better?"
                    className="form-control feedback-prompt__textarea"
                />
                <Button className="w-100">Send</Button>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
