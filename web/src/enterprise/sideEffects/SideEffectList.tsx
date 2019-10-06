import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useState } from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ConnectionListHeaderItems } from '../../components/connectionList/ConnectionListHeader'
import { SideEffectListItem, SideEffectListItemContext } from './SideEffectListItem'

const LOADING: 'loading' = 'loading'

export interface SideEffectListContext extends SideEffectListItemContext {
    history: H.History
    location: H.Location
}

interface Props extends SideEffectListContext {
    sideEffects: typeof LOADING | GQL.ISideEffectConnection | ErrorLike

    headerItems?: ConnectionListHeaderItems

    className?: string
}

/**
 * A list of side effects.
 */
export const SideEffectList: React.FunctionComponent<Props> = ({ sideEffects, className = '', ...props }) => {
    const DEFAULT_MAX = 12
    const [max, setMax] = useState<number | undefined>(DEFAULT_MAX)
    const onShowAllClick = useCallback(() => setMax(undefined), [])

    return (
        <div className={`side-effect-list ${className}`}>
            {isErrorLike(sideEffects) ? (
                <div className="alert alert-danger">{sideEffects.message}</div>
            ) : (
                <div className="card">
                    <h4 className="card-header">Side effects</h4>
                    {sideEffects === LOADING ? (
                        <LoadingSpinner className="m-3" />
                    ) : sideEffects.nodes.length === 0 ? (
                        <p className="p-3 mb-0 text-muted">No sideEffects.</p>
                    ) : (
                        <>
                            <ul className="list-group list-group-flush">
                                {sideEffects.nodes.slice(0, max).map((node, i) => (
                                    <SideEffectListItem key={i} {...props} sideEffect={node} />
                                ))}
                            </ul>
                            {max !== undefined && sideEffects.nodes.length > max && (
                                <div className="card-footer p-0">
                                    <button type="button" className="btn btn-sm btn-link" onClick={onShowAllClick}>
                                        Show {sideEffects.nodes.length - max} more
                                    </button>
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    )
}
