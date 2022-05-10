import React from 'react'

import { Typography } from '@sourcegraph/wildcard'

import styles from './EmptyChangesetListElement.module.scss'

export const EmptyChangesetListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className={styles.emptyChangesetListElementBody}>
        <Typography.H2 className="text-center mb-4">This batch change does not contain changesets</Typography.H2>
        <p>This can occur for several reasons:</p>
        <p>
            <strong>
                The query specified in <span className="text-monospace">repositoriesMatchingQuery:</span> may not have
                matched any repositories.
            </strong>
        </p>
        <p>Test your query in the search bar and ensure it returns results.</p>
        <p>
            <strong>
                The code specified in <span className="text-monospace">steps:</span> may not have resulted in changes
                being made.
            </strong>
        </p>
        <p>
            Try the command on a local instance of one of the repositories returned in your search results. Run{' '}
            <span className="text-monospace">git status</span> and ensure it produced changed files.
        </p>
    </div>
)
