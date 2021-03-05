import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

interface Props {
    /** used to build the key that represents the alert in local storage */
    partialStorageKey: string

    /** class name to be applied to the alert */
    className: string
}

/**
 * A global site alert that can be dismissed. Once dismissed, it is never shown
 * again.
 */
export const DismissibleAlert: React.FunctionComponent<Props> = ({ partialStorageKey, className, children }) => {
    const key = `DismissibleAlert/${partialStorageKey}/dismissed`
    const [dismissed, setDismissed] = React.useState<boolean>(localStorage.getItem(key) === 'true')

    const onDismiss = React.useCallback(() => {
        localStorage.setItem(key, 'true')
        setDismissed(true)
    }, [key])

    if (dismissed) {
        return null
    }
    return (
        <div className={`alert dismissible-alert ${className}`}>
            <div className="dismissible-alert__content">{children}</div>
            <button type="button" className="btn btn-icon" onClick={onDismiss}>
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}
