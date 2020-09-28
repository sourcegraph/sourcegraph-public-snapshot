import { storiesOf } from '@storybook/react'
import React, { useCallback, useState } from 'react'
import { ModalContainer } from './ModalContainer'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/ModalContainer', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('Interactive + Centered', () => {
    const [open, setOpen] = useState(true)
    const onClose = useCallback(() => setOpen(false), [])
    const openModal = useCallback(() => setOpen(true), [])

    return (
        <WebStory>
            {webProps => (
                <div>
                    <button type="button" className="btn btn-primary" onClick={openModal}>
                        Open modal
                    </button>
                    {open && (
                        <ModalContainer
                            {...webProps}
                            onClose={onClose}
                            hideCloseIcon={true}
                            className="justify-content-center"
                        >
                            {bodyReference => (
                                <div
                                    className="extension-permission-modal p-4"
                                    ref={bodyReference as React.MutableRefObject<HTMLDivElement>}
                                >
                                    <h1>Modal</h1>
                                    <p>You can click outside of the modal body to close me</p>
                                </div>
                            )}
                        </ModalContainer>
                    )}
                </div>
            )}
        </WebStory>
    )
})
