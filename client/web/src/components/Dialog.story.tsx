import { storiesOf } from '@storybook/react'
import React, { useCallback, useState } from 'react'
import { Dialog } from './Dialog'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/Dialog', module).addDecorator(story => <div className="container mt-3">{story()}</div>)

add('Interactive', () => {
    const [open, setOpen] = useState(true)
    const onClose = useCallback(() => setOpen(false), [])
    const openDialog = useCallback(() => setOpen(true), [])

    return (
        <WebStory>
            {() => (
                <div>
                    <button type="button" className="btn btn-primary" onClick={openDialog}>
                        Open modal
                    </button>
                    {open && (
                        <Dialog onClose={onClose}>
                            <div className="extension-permission-modal p-4">
                                <h1>Dialog</h1>
                                <p>You can click on the backdrop to close me</p>
                            </div>
                        </Dialog>
                    )}
                </div>
            )}
        </WebStory>
    )
})
