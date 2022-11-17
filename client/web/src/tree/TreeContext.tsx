import React, { useContext } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { AbsoluteRepo } from '@sourcegraph/shared/src/util/url'

export interface TreeRootContext extends AbsoluteRepo {
    rootTreeUrl: string
    repoID: Scalars['ID']
}

export const TreeRootContext = React.createContext<TreeRootContext>({
    rootTreeUrl: '',
    repoID: '',
    repoName: '',
    revision: '',
    commitID: '',
})

export const useTreeRootContext = (): TreeRootContext => useContext(TreeRootContext)
