import { ChevronDownIcon } from 'mdi-react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { TreeLayerProps } from './TreeLayer'
import { treePadding } from './util'

interface TreeChildProps extends TreeLayerProps {
    className: string
    maxEntries: number
    handleTreeClick: () => void
    noopRowClick: (e: React.MouseEvent<HTMLAnchorElement>) => void
    linkRowClick: (e: React.MouseEvent<HTMLAnchorElement>) => void
}

/**
 * JSX to render a tree directory
 * @param props
 */
export const Directory: React.SFC<TreeChildProps> = (props: TreeChildProps): JSX.Element => (
    <tr key={props.entryInfo.path} className={props.className} onClick={props.handleTreeClick}>
        <td className="tree__cell">
            <div
                className="tree__row-contents tree__row-contents-new"
                data-tree-is-directory="true"
                data-tree-path={props.entryInfo.path}
            >
                <div className="tree__row-contents-text">
                    <a
                        className="tree__row-icon"
                        href={props.entryInfo.url}
                        onClick={props.noopRowClick}
                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                        style={treePadding(props.depth, true)}
                        tabIndex={-1}
                    >
                        {props.isExpanded ? (
                            <ChevronDownIcon className="icon-inline" />
                        ) : (
                            <ChevronRightIcon className="icon-inline" />
                        )}
                    </a>
                    <Link
                        to={props.entryInfo.url}
                        onClick={props.linkRowClick}
                        className="tree__row-label"
                        draggable={false}
                        title={props.entryInfo.path}
                        tabIndex={-1}
                    >
                        {props.entryInfo.name}
                    </Link>
                </div>
            </div>
            {props.index === props.maxEntries - 1 && (
                <div
                    className="tree__row-alert alert alert-warning"
                    // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                    style={treePadding(props.depth, true)}
                >
                    Too many entries. Use search to find a specific file.
                </div>
            )}
        </td>
    </tr>
)
