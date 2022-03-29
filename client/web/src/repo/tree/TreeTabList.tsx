import React from 'react'

import BookOpenBlankVariantIcon from 'mdi-react/BookOpenBlankVariantIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'

import { Icon, Link } from '@sourcegraph/wildcard'

import { TreeFields } from '../../graphql-operations'

interface TreeTabList {
    tree: TreeFields
}

export const TreeTabList: React.FunctionComponent<TreeTabList> = ({ tree }) => (
    <div>
        <Link to={`${tree.url}/-/docs/tab`}>
            <Icon as={BookOpenBlankVariantIcon} /> API docs
        </Link>
        <Link to={`${tree.url}/-/commits/tab`}>
            <Icon as={SourceCommitIcon} /> Commits
        </Link>
    </div>
)
