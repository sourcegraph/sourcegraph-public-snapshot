import React, { useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { ExtensionStatus } from '../../../shared/src/extensions/ExtensionStatus'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { Modal } from 'reactstrap'

interface Props
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'sideloadedExtensionURL'> {
    className?: string
}

export const SHOW_DEBUG = localStorage.getItem('debug') !== null

const ExtensionLink: React.FunctionComponent<{ id: string }> = ({ id }) => <Link to={`/extensions/${id}`}>{id}</Link>

/**
 * A button that toggles the visibility of the global debug modal. It should only be shown if
 * `SHOW_DEBUG` is true.
 */
export const GlobalDebugModalButton: React.FunctionComponent<Props> = ({ ...props }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => {
        setIsOpen(!isOpen)
    }, [isOpen])

    return (
        <>
            <button type="button" className={props.className} onClick={toggleIsOpen} data-tooltip="Developer console">
                Dev
            </button>
            <Modal isOpen={isOpen} toggle={toggleIsOpen} centered={true} fade={false}>
                <ExtensionStatus {...props} link={ExtensionLink} />
            </Modal>
        </>
    )
}
