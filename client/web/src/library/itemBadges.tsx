import type { FunctionComponent } from 'react'

import classNames from 'classnames'

import { Badge } from '@sourcegraph/wildcard'

import { SavedSearchVisibility, type PromptFields, type SavedSearchFields } from '../graphql-operations'

type LibraryItem = SavedSearchFields | PromptFields

export const LibraryItemStatusBadge: FunctionComponent<{
    item: Pick<LibraryItem, 'draft'>
    className?: string
}> = ({ item: { draft }, className }) =>
    draft ? (
        <Badge variant="outlineSecondary" small={true} className={classNames('font-italic', className)}>
            Draft
        </Badge>
    ) : null

export const LibraryItemVisibilityBadge: FunctionComponent<{
    item: Pick<LibraryItem, '__typename' | 'visibility' | 'owner'>
    className?: string
}> = ({ item, className }) =>
    item.visibility === SavedSearchVisibility.SECRET ? (
        <Badge
            variant="outlineSecondary"
            small={true}
            className={className}
            tooltip={[
                item.owner.__typename === 'User'
                    ? `Only ${
                          item.owner.id === window.context?.currentUser?.id ? 'you' : 'the user who owns it'
                      } can see this ${itemNoun(item)}. Transfer it to an organization to share it with other users.`
                    : `Only members of the "${item.owner.namespaceName}" organization can see this ${itemNoun(item)}.`,
                'Ask a site admin to make it public if you want all users to be able to view it.',
            ].join(' ')}
        >
            Secret
        </Badge>
    ) : (
        <Badge
            variant="outlineSecondary"
            small={true}
            className={className}
            tooltip={`All users can view this ${itemNoun(item)}.`}
        >
            Public
        </Badge>
    )

function itemNoun(item: Pick<LibraryItem, '__typename'>): string {
    return item.__typename === 'SavedSearch' ? 'saved search' : item.__typename === 'Prompt' ? 'prompt' : 'item'
}
