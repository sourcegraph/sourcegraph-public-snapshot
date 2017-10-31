
import LockIcon from '@sourcegraph/icons/lib/Lock'
import UnlockIcon from '@sourcegraph/icons/lib/Unlock'
import * as React from 'react'

export function SecurityWidget(sharedItem: GQL.ISharedItem): JSX.Element | null {
    if (sharedItem.public) {
        return (
            <div className='security-widget'>
                <div className='security-widget__main-label'>
                    <UnlockIcon className='icon-inline' /> Secret URL
                </div>
                <div className='security-widget__extended-help'>
                    Anyone with the link can view this page.
                </div>
            </div>
        )
    } else {
        return (
            <div className='security-widget'>
                <div className='security-widget__main-label'>
                    <LockIcon className='icon-inline' /> Organization Only
                </div>
                <div className='security-widget__extended-help'>
                    Only members of your organization can view this page.
                </div>
            </div>
        )
    }
}
