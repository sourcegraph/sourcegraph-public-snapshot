import * as React from 'react'
import { Subscription } from 'rxjs'
import * as permissions from '../../../browser/permissions'
import storage from '../../../browser/storage'
import { StorageItems } from '../../../browser/types'
import { GQL } from '../../../types/gqlschema'
import { fetchCurrentUser } from '../../backend/server'
import { ConnectionCard } from './ConnectionCard'
import { FeatureFlagCard } from './FeatureFlagCard'
import { SettingsCard } from './SettingsCard'

interface Props {}
interface State {
    currentUser: GQL.IUser | undefined
    storage: StorageItems | undefined
    permissionOrigins: string[]
}

/**
 * A page displaying an overview of the extension configuration state.
 */
export class OptionsConfiguration extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            storage: undefined,
            currentUser: undefined,
            permissionOrigins: [],
        }
    }

    public componentDidMount(): void {
        fetchCurrentUser(true).subscribe(user => {
            this.setState(() => ({ currentUser: user }))
        })
        storage.onChanged(() => {
            this.updateForStorageItems()
        })
        permissions.onAdded(() => {
            this.updateForPermissions()
        })
        permissions.onRemoved(() => {
            this.updateForPermissions()
        })
        this.updateForStorageItems()
        this.updateForPermissions()
    }

    private updateForStorageItems = () => {
        storage.getSync(items => {
            this.setState(() => ({ storage: items }))
        })
    }

    private updateForPermissions = () => {
        permissions.getAll().then(
            permissions => {
                this.setState(() => ({ permissionOrigins: permissions.origins || [] }))
            },
            () => {
                /** noop */
            }
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { storage, currentUser, permissionOrigins } = this.state
        if (!storage) {
            return null
        }
        return (
            <div className="options-configuration-page">
                <ConnectionCard permissionOrigins={permissionOrigins} storage={storage} currentUser={currentUser} />
                <FeatureFlagCard storage={storage} />
                {storage.useExtensions && <SettingsCard currentUser={currentUser} />}
            </div>
        )
    }
}
