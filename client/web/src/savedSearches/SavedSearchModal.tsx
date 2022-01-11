import Dialog from '@reach/dialog'
import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'

import { AuthenticatedUser } from '../auth'
import { SearchPatternTypeProps } from '../search'

import styles from './SavedSearchModal.module.scss'

interface Props extends SearchPatternTypeProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
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

const MODAL_LABEL_ID = 'saved-search-modal-id'

export class SavedSearchModal extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            saveLocation: UserOrOrg.User,
        }
    }

    private onLocationChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        const locationType = event.target.value
        this.setState({ saveLocation: locationType as UserOrOrg })
    }

    private onOrganizationChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        const orgName = event.target.value
        this.setState({ organization: orgName })
    }

    public render(): JSX.Element | null {
        return (
            this.props.authenticatedUser && (
                <Dialog
                    aria-labelledby={MODAL_LABEL_ID}
                    className={styles.savedSearchModalForm}
                    onDismiss={this.props.onDidCancel}
                    data-testid="saved-search-modal"
                >
                    <Form onSubmit={this.onSubmit} className="test-saved-search-modal">
                        <h3 id={MODAL_LABEL_ID}>Save search query to: </h3>
                        <div className="form-group">
                            <select
                                onChange={this.onLocationChange}
                                className={classNames(styles.select, 'form-control')}
                            >
                                <option value={UserOrOrg.User}>User</option>
                                {this.props.authenticatedUser.organizations &&
                                    this.props.authenticatedUser.organizations.nodes.length > 0 && (
                                        <option value={UserOrOrg.Org}>Organization</option>
                                    )}
                            </select>
                            {this.props.authenticatedUser.organizations &&
                                this.props.authenticatedUser.organizations.nodes.length > 0 &&
                                this.state.saveLocation === UserOrOrg.Org && (
                                    <select
                                        onChange={this.onOrganizationChange}
                                        placeholder="Select an organization"
                                        className={classNames(styles.select, 'form-control')}
                                    >
                                        <option value="" disabled={true} selected={true}>
                                            Select an organization
                                        </option>
                                        {this.props.authenticatedUser.organizations.nodes.map(org => (
                                            <option value={org.name} key={org.name}>
                                                {org.displayName ? org.displayName : org.name}
                                            </option>
                                        ))}
                                    </select>
                                )}
                        </div>
                        <button
                            type="submit"
                            disabled={this.state.saveLocation === UserOrOrg.Org && !this.state.organization}
                            className={classNames(styles.button, 'btn btn-primary test-saved-search-modal-save-button')}
                        >
                            Save query
                        </button>
                    </Form>
                </Dialog>
            )
        )
    }

    private onSubmit = (): void => {
        if (this.props.query && this.props.authenticatedUser) {
            const encodedQuery = encodeURIComponent(this.props.query)
            this.props.history.push(
                this.state.saveLocation.toLowerCase() === 'user'
                    ? `/users/${this.props.authenticatedUser.username}/searches/add?query=${encodedQuery}&patternType=${this.props.patternType}`
                    : `/organizations/${this.state.organization!}/searches/add?query=${encodedQuery}&patternType=${
                          this.props.patternType
                      }`
            )
        }
    }
}
