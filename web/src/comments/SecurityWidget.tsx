import LockIcon from '@sourcegraph/icons/lib/Lock'
import UnlockIcon from '@sourcegraph/icons/lib/Unlock'
import * as React from 'react'

interface Props {
    sharedItem: GQL.ISharedItem
}

export function SecurityWidget(props: Props): JSX.Element | null {
    if (props.sharedItem.public) {
        return (
            <div className="security-widget">
                <div className="security-widget__main-label">
                    <UnlockIcon className="icon-inline" /> Secret URL
                </div>
                <div className="security-widget__extended-help">Anyone with the link can view this page.</div>
            </div>
        )
    } else {
        return (
            <div className="security-widget">
                <div className="security-widget__main-label">
                    <LockIcon className="icon-inline" /> Organization Only
                </div>
                <div className="security-widget__extended-help">
                    Only members of your organization can view this page.
                </div>
            </div>
        )
    }
}
