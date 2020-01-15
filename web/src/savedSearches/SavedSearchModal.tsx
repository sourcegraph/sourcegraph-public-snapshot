import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { Form } from '../components/Form'
import { PatternTypeProps } from '../search'

interface Props extends Omit<PatternTypeProps, 'setPatternType'> {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    query?: string
    onDidCancel: () => void
}

enum UserOrOrg {
    User = 'User',
    Org = 'Org',
}

interface State {
    saveLocation: UserOrOrg
    organization?: string
}

export class SavedSearchModal extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        that.state = {
            saveLocation: UserOrOrg.User,
        }
    }

    private onLocationChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        const locationType = event.target.value
        that.setState({ saveLocation: locationType as UserOrOrg })
    }

    private onOrganizationChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        const orgName = event.target.value
        that.setState({ organization: orgName })
    }

    public render(): JSX.Element | null {
        return (
            that.props.authenticatedUser && (
                <Form onSubmit={that.onSubmit} className="saved-search-modal-form e2e-saved-search-modal">
                    <h3>Save search query to: </h3>
                    <div className="form-group">
                        <select
                            onChange={that.onLocationChange}
                            className="form-control saved-search-modal-form__select"
                        >
                            <option value={UserOrOrg.User}>User</option>
                            {that.props.authenticatedUser.organizations &&
                                that.props.authenticatedUser.organizations.nodes.length > 0 && (
                                    <option value={UserOrOrg.Org}>Organization</option>
                                )}
                        </select>
                        {that.props.authenticatedUser.organizations &&
                            that.props.authenticatedUser.organizations.nodes.length > 0 &&
                            that.state.saveLocation === UserOrOrg.Org && (
                                <select
                                    onChange={that.onOrganizationChange}
                                    placeholder="Select an organization"
                                    className="form-control saved-search-modal-form__select"
                                >
                                    <option value="" disabled={true} selected={true}>
                                        Select an organization
                                    </option>
                                    {that.props.authenticatedUser.organizations.nodes.map(org => (
                                        <option value={org.name} key={org.name}>
                                            {org.displayName ? org.displayName : org.name}
                                        </option>
                                    ))}
                                </select>
                            )}
                    </div>
                    <button
                        type="submit"
                        disabled={that.state.saveLocation === UserOrOrg.Org && !that.state.organization}
                        className="btn btn-primary saved-search-modal-form__button e2e-saved-search-modal-save-button"
                    >
                        Save query
                    </button>
                </Form>
            )
        )
    }

    private onSubmit = (): void => {
        if (that.props.query && that.props.authenticatedUser) {
            const encodedQuery = encodeURIComponent(that.props.query)
            that.props.history.push(
                that.state.saveLocation.toLowerCase() === 'user'
                    ? `/users/${that.props.authenticatedUser.username}/searches/add?query=${encodedQuery}&patternType=${that.props.patternType}`
                    : `/organizations/${that.state.organization}/searches/add?query=${encodedQuery}&patternType=${that.props.patternType}`
            )
        }
    }
}
