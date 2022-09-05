import React from 'react'

import { H3, Text } from '@sourcegraph/wildcard'

import styles from './EmptyPreviewListElement.module.scss'

export const EmptyPreviewListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className={styles.emptyPreviewListElementBody}>
        <H3 className="text-center mb-4">No changesets will be created by this batch change</H3>
        <Text>This can occur for several reasons:</Text>
        <Text>
            <strong>
                The query specified in <span className="text-monospace">repositoriesMatchingQuery:</span> may not have
                matched any repositories.
            </strong>
        </Text>
        <Text>Test your query in the search bar and ensure it returns results.</Text>
        <Text>
            <strong>
                The code specified in <span className="text-monospace">steps:</span> may not have resulted in changes
                being made.
            </strong>
        </Text>
        <Text>
            Try the command on a local instance of one of the repositories returned in your search results. Run{' '}
            <span className="text-monospace">git status</span> and ensure it produced changed files.
        </Text>
    </div>
)
