import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { FeedbackBadge, PageHeader } from '@sourcegraph/wildcard'

import { SearchStreamingProps } from '..'
import { fetchRepository, resolveRevision } from '../../repo/backend'
import { StreamingSearchResultsListProps } from '../results/StreamingSearchResultsList'

import { SearchNotebook } from './SearchNotebook'
import styles from './SearchNotebookPage.module.scss'
import { serializeBlocks, deserializeBlockInput } from './serialize'

import { Block, BlockInput } from '.'

interface SearchNotebookPageProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded'>,
        ExtensionsControllerProps<'extHostAPI'> {
    globbing: boolean
    isMacPlatform: boolean
}

export const SearchNotebookPage: React.FunctionComponent<SearchNotebookPageProps> = props => {
    useEffect(() => props.telemetryService.logViewEvent('SearchNotebookPage'), [props.telemetryService])

    const history = useHistory()
    const location = useLocation()

    const onSerializeBlocks = useCallback(
        (blocks: Block[]) => history.replace({ hash: serializeBlocks(blocks, window.location.origin) }),
        [history]
    )

    const blocks: BlockInput[] = useMemo(() => {
        const serializedBlocks = location.hash.slice(1)
        if (serializedBlocks.length === 0) {
            return [
                { type: 'md', input: '## Welcome to Sourcegraph Search Notebooks!\nDo something awesome.' },
                { type: 'query', input: '// Enter a search query' },
            ]
        }

        return serializedBlocks.split(',').map(serializedBlock => {
            const [type, encodedInput] = serializedBlock.split(':')
            if (type === 'md' || type === 'query' || type === 'file') {
                return deserializeBlockInput(type, decodeURIComponent(encodedInput))
            }
            throw new Error(`Unknown block type: ${type}`)
        })
        // Deserialize only on startup
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return (
        <div className={classNames('w-100 p-2', styles.searchNotebookPage)}>
            <PageTitle title="Search Notebook" />
            <Page>
                <PageHeader
                    className="mx-3"
                    annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                    path={[{ text: 'Search Notebook' }]}
                />
                <hr className="mt-2 mb-1 mx-3" />
                <SearchNotebook
                    {...props}
                    blocks={blocks}
                    onSerializeBlocks={onSerializeBlocks}
                    resolveRevision={resolveRevision}
                    fetchRepository={fetchRepository}
                />
            </Page>
        </div>
    )
}
