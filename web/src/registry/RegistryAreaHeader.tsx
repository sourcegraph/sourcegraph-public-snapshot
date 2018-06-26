import AddIcon from '@sourcegraph/icons/lib/Add'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { RegistryAreaPageProps } from './RegistryArea'

interface Props extends RegistryAreaPageProps, RouteComponentProps<{}> {
    className: string

    showActions: 'primary' | 'link' | false
    mainHeader: boolean
}

const Header: React.SFC<{ mainHeader: boolean; children: JSX.Element }> = ({ mainHeader, children }) =>
    mainHeader ? <h2 className="mb-2">{children}</h2> : <h4 className="mb-0">{children}</h4>

/**
 * Header for the registry area.
 */
export const RegistryAreaHeader: React.SFC<Props> = (props: Props) => (
    <div className={`area-header ${props.className} ${props.mainHeader ? '' : 'py-1'}`}>
        <div className={`${props.className}-inner d-flex justify-content-between align-items-center`}>
            <Header mainHeader={props.mainHeader}>
                <Link className={props.mainHeader ? 'registry-area__title' : ''} to="/registry">
                    {props.mainHeader && <PuzzleIcon className="icon-inline" />} Extension registry
                </Link>
            </Header>
            {props.showActions !== false && (
                <div>
                    {!props.mainHeader && (
                        <Link
                            className={`btn mx-2 btn-${props.showActions} ${props.mainHeader ? '' : 'btn-sm'}`}
                            to="/registry"
                        >
                            View all extensions
                        </Link>
                    )}
                    <Link
                        className={`btn ml-2 btn-${props.showActions} ${props.mainHeader ? '' : 'btn-sm'}`}
                        to="/registry/extensions/new"
                    >
                        {props.mainHeader && <AddIcon className="icon-inline" />} Create new extension
                    </Link>
                </div>
            )}
        </div>
    </div>
)
