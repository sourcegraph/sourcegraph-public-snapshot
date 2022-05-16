import React from 'react'

import { splitExtensionID } from '@sourcegraph/shared/src/extensions/extension'
import { Button, Modal, Typography } from '@sourcegraph/wildcard'

/**
 * A modal confirmation prompt to the user confirming whether to add an extension.
 */
export const ExtensionPermissionModal: React.FunctionComponent<
    React.PropsWithChildren<{
        extensionID: string
        givePermission: () => void
        denyPermission: () => void
    }>
> = ({ extensionID, denyPermission, givePermission }) => {
    const { name } = splitExtensionID(extensionID)
    const labelId = `label--permission-${extensionID}`

    return (
        <Modal position="center" onDismiss={denyPermission} aria-labelledby={labelId}>
            <Typography.H3 id={labelId}>Add {name || extensionID} Sourcegraph extension?</Typography.H3>
            <p className="mb-0 mt-3">It will be able to:</p>
            <ul className="list-dashed">
                <li className="m-0">read repositories and files you view using Sourcegraph</li>
                <li className="m-0">read and change your Sourcegraph settings</li>
            </ul>
            <div className="d-flex justify-content-end pt-5">
                <Button className="mr-2" onClick={denyPermission} outline={true} variant="secondary">
                    Cancel
                </Button>
                <Button onClick={givePermission} variant="primary">
                    Yes, add {name || extensionID}!
                </Button>
            </div>
        </Modal>
    )
}
