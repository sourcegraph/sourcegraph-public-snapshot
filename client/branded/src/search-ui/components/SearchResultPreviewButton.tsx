import { FC } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { useSearchResultState, type SearchResultPreview } from '../stores/results-store'

interface SearchResultPreviewButtonProps {
    result: SearchResultPreview
}

export const SearchResultPreviewButton: FC<SearchResultPreviewButtonProps> = props => {
    const { result } = props
    const { previewBlob, setPreviewBlob, clearPreview } = useSearchResultState()

    const isActive =
        previewBlob?.repository === result.repository &&
        previewBlob?.path === result.path &&
        previewBlob.commit === result.commit

    const handleClick = (): void => {
        if (isActive) {
            clearPreview()
        } else {
            setPreviewBlob(result)
        }
    }

    return (
        <Button variant="link" className="py-1 px-0 mr-2" onClick={handleClick}>
            {isActive ? 'Hide preview' : 'Preview'}
        </Button>
    )
}
