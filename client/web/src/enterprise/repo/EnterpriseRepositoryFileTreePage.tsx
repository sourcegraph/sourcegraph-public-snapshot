import type { FC } from 'react'

import { type RepositoryFileTreePageProps, RepositoryFileTreePage } from '../../repo/RepositoryFileTreePage'
import { useCodeIntel } from '../codeintel/useCodeIntel'

export const EnterpriseRepositoryFileTreePage: FC<RepositoryFileTreePageProps> = props => (
    <RepositoryFileTreePage {...props} useCodeIntel={useCodeIntel} />
)
