import * as React from 'react'

import { Link, Alert } from '@sourcegraph/wildcard'

export const DirectImportRepoAlert: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className = '',
}) => (
    <>
        {['dev', 'docker-container'].includes(window.context.deployType) && (
            <Alert className={className} variant="info">
                Very large repository? See{' '}
                <Link to="/help/admin/repo/pre_load_from_local_disk#add-repositories-already-cloned-to-disk">
                    how to reuse an existing local clone
                </Link>{' '}
                to speed this up.
            </Alert>
        )}
    </>
)
