import { noop } from 'lodash'
import React, { useMemo } from 'react'

import { convertMarkdownToBlocks } from './convertMarkdownToBlocks'
import { SearchNotebook, SearchNotebookProps } from './SearchNotebook'

export const SEARCH_NOTEBOOK_FILE_EXTENSION = '.snb.md'

interface RenderedSearchNotebookMarkdownProps extends Omit<SearchNotebookProps, 'onSerializeBlocks' | 'blocks'> {
    markdown: string
}

export const RenderedSearchNotebookMarkdown: React.FunctionComponent<RenderedSearchNotebookMarkdownProps> = ({
    markdown,
    ...props
}) => {
    const blocks = useMemo(() => convertMarkdownToBlocks(markdown), [markdown])
    return <SearchNotebook isReadOnly={true} blocks={blocks} {...props} onSerializeBlocks={noop} />
}
