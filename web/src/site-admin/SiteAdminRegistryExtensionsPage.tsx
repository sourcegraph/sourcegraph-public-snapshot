import AddIcon from '@sourcegraph/icons/lib/Add'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { ExtensionsListViewMode, RegistryExtensionsList } from '../registry/RegistryExtensionsPage'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser | null
}

/** Displays all registry extensions on this site. */
export class SiteAdminRegistryExtensionsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminRegistryExtensions')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extensions-page">
                <PageTitle title="Registry extensions" />
                <div className="d-flex justify-content-between align-items-center">
                    <h2 className="mr-sm-2 mb-0">Registry extensions</h2>
                    <div>
                        <Link className="btn btn-outline-link mr-sm-2" to="/registry">
                            View extension registry
                        </Link>
                        <Link className="btn btn-primary" to="/registry/extensions/new">
                            <AddIcon className="icon-inline" /> Create new extension
                        </Link>
                    </div>
                </div>
                <p className="mt-2">
                    Extensions add features to Sourcegraph and other connected tools (such as editors, code hosts, and
                    code review tools).
                </p>
                <RegistryExtensionsList
                    {...this.props}
                    authenticatedUser={this.props.user}
                    mode={ExtensionsListViewMode.List}
                    publisher={null}
                    showExtensionID="extensionID"
                    showDeleteAction={true}
                    showEditAction={true}
                    showTimestamp={true}
                    filters={RegistryExtensionsList.FILTERS}
                />
            </div>
        )
    }
}
