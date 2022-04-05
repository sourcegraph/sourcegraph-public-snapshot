import React from 'react'

import { isErrorLike } from '@sourcegraph/common'

import { RepositoryMenuContentProps } from '../../codeintel/RepositoryMenu'

export const RepositoryMenuContent: React.FunctionComponent<RepositoryMenuContentProps> = props => {
    const forNerds =
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final?.experimentalFeatures?.codeIntelRepositoryBadge?.forNerds

    return (
        <>
            <div className="px-2 py-1">
                <h2>Unimplemented</h2>

                <p className="text-muted">Unimplemented (enterprise version).</p>
            </div>

            {forNerds && <div className="px-2 py-1">NERD DATA</div>}
        </>
    )
}
