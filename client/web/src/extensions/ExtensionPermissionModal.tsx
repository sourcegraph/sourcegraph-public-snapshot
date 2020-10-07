import React from 'react'
import { ModalContainer } from '../components/ModalContainer'
import { splitExtensionID } from './extension/extension'

/**
 * A modal confirmation prompt to the user confirming whether to add an extension.
 */
export const ExtensionPermissionModal: React.FunctionComponent<{
    extensionID: string
    givePermission: () => void
    denyPermission: () => void
}> = ({ extensionID, denyPermission, givePermission }) => {
    const { name } = splitExtensionID(extensionID)

    return (
        <ModalContainer className="justify-content-center" onClose={denyPermission} hideCloseIcon={true}>
            {bodyReference => (
                <div
                    className="extension-permission-modal p-4"
                    ref={bodyReference as React.MutableRefObject<HTMLDivElement>}
                >
                    <h3>Add {name || extensionID} Sourcegraph extension?</h3>
                    <p className="mb-0 mt-3">It will be able to:</p>
                    <ul className="list-dashed">
                        <li className="m-0">read repositories and files you view using Sourcegraph</li>
                        <li className="m-0">read and change your Sourcegraph settings</li>
                    </ul>
                    <div className="d-flex justify-content-end pt-5">
                        <button type="button" className="btn btn-outline-secondary mr-2" onClick={denyPermission}>
                            Cancel
                        </button>
                        <button type="button" className="btn btn-primary" onClick={givePermission}>
                            Yes, add {name || extensionID}!
                        </button>
                    </div>
                </div>
            )}
        </ModalContainer>
    )
}
