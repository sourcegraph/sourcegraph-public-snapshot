import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { ActionButtonDescriptor } from '../util/contributions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { PageHeader } from '../components/PageHeader'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

export interface ExtensionsAreaHeaderProps extends ExtensionsAreaRouteContext, RouteComponentProps<{}> {
    isPrimaryHeader: boolean
    actionButtons: readonly ExtensionsAreaHeaderActionButton[]
}

export interface ExtensionAreaHeaderContext {
    isPrimaryHeader: boolean
}

export interface ExtensionsAreaHeaderActionButton extends ActionButtonDescriptor<ExtensionAreaHeaderContext> {}

/**
 * Header for the extensions area.
 */
export const ExtensionsAreaHeader: React.FunctionComponent<ExtensionsAreaHeaderProps> = (
    props: ExtensionsAreaHeaderProps
) => (
    <div className="container">
        {props.isPrimaryHeader && (
            <PageHeader
                title="Extensions"
                icon={PuzzleOutlineIcon}
                actions={props.actionButtons.map(
                    ({ condition = () => true, to, icon: Icon, label, tooltip }) =>
                        condition(props) && (
                            <Link className="btn ml-2 btn-secondary" to={to(props)} data-tooltip={tooltip}>
                                {Icon && <Icon className="icon-inline" />} {label}
                            </Link>
                        )
                )}
            />
        )}
    </div>
)
