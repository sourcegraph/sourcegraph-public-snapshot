import React from 'react'
import { ModalContainer } from '../components/ModalContainer'

/**
 * A modal confirmation prompt to the user confirming whether to add an extension.
 */
export const ExtensionPermissionModal: React.FunctionComponent<{
    extensionID: string
    givePermission: () => void
    denyPermission: () => void
}> = ({ extensionID, denyPermission, givePermission }) => {
    const extensionName = extensionID.split('/')[1]

    return (
        <ModalContainer
            component={
                <div className="extension-permission-modal p-4">
                    <h3>Add {extensionName || extensionID} Sourcegraph extension?</h3>
                    <p className="mb-0 mt-3">It will be able to:</p>
                    <p className="m-0">- read repositories and files you view using Sourcegraph</p>
                    <p className="m-0">- read and change your Sourcegraph settings</p>
                    <div className="d-flex justify-content-end pt-5">
                        <button type="button" className="btn btn-outline-secondary mr-2" onClick={denyPermission}>
                            No
                        </button>
                        <button type="button" className="btn btn-primary" onClick={givePermission}>
                            Yes, add {extensionName || extensionID}!
                        </button>
                    </div>
                </div>
            }
        />
    )
}
