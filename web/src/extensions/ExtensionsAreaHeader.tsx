import AddIcon from '@sourcegraph/icons/lib/Add'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { ExtensionsAreaPageProps } from './ExtensionsArea'

interface Props extends ExtensionsAreaPageProps, RouteComponentProps<{}> {
    isPrimaryHeader: boolean
}

/**
 * Header for the extensions area.
 */
export const ExtensionsAreaHeader: React.SFC<Props> = (props: Props) => (
    <div className="border-bottom simple-area-header">
        <div className="navbar navbar-expand py-2">
            <div className="container">
                {props.isPrimaryHeader && (
                    <h3 className="mb-0">
                        <Link className="nav-brand mr-2" to="/extensions">
                            <strong>Extensions</strong>
                        </Link>
                    </h3>
                )}
                <ul className="navbar-nav nav">
                    {!props.isPrimaryHeader && (
                        <li className="nav-item">
                            <Link to="/extensions" className="nav-link">
                                <ArrowLeftIcon className="icon-inline" /> All extensions
                            </Link>
                        </li>
                    )}
                </ul>
                <div className="spacer" />
                <ul className="navbar-nav nav">
                    {props.isPrimaryHeader && (
                        <li className="nav-item">
                            <Link className="btn ml-2 btn-secondary" to="/extensions/registry/new">
                                {props.isPrimaryHeader && <AddIcon className="icon-inline" />} Publish new extension
                            </Link>
                        </li>
                    )}
                </ul>
            </div>
        </div>
    </div>
)
