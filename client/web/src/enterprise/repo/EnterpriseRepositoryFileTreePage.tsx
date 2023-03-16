import { FC } from 'react'

import { RepositoryFileTreePageProps, RepositoryFileTreePage } from '../../repo/RepositoryFileTreePage'
import { useCodeIntel } from '../codeintel/useCodeIntel'

export const EnterpriseRepositoryFileTreePage: FC<RepositoryFileTreePageProps> = props => (
    <RepositoryFileTreePage {...props} useCodeIntel={useCodeIntel} />
)
