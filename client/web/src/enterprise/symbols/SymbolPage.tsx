import React, { useEffect, useMemo, useState } from 'react'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'

export interface SymbolRouteProps extends Pick<RepoRevisionContainerContext, 'repo' | 'revision'> {}

export const SymbolPage: React.FunctionComponent<SymbolRouteProps> = ({ repo, revision, ...props }) => {
    return <div>Symbol page</div>
}
