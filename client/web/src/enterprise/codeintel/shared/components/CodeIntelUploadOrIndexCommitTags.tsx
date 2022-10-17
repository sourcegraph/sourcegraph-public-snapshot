import React, { FunctionComponent } from 'react'

import { Badge } from '@sourcegraph/wildcard'

export interface CodeIntelUploadOrIndexCommitTagsProps {
    tags: string[]
}

export const CodeIntelUploadOrIndexCommitTags: FunctionComponent<
    React.PropsWithChildren<CodeIntelUploadOrIndexCommitTagsProps>
> = ({ tags }) => (
    <>
        {tags.length > 0 && (
            <>
                tagged as{' '}
                {tags
                    .slice(0, 3)
                    .map<React.ReactNode>(tag => (
                        <Badge key={tag} variant="outlineSecondary">
                            {tag}
                        </Badge>
                    ))
                    .reduce((previous, current) => [previous, ', ', current])}
                {tags.length > 3 && <> and {tags.length - 3} more</>}
            </>
        )}
    </>
)
