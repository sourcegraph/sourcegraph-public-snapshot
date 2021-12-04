import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogEntityCodeOwnersFields } from '../../../../../graphql-operations'

import { PersonListRow } from './PersonListRow'

interface Props {
    entity: CatalogEntityCodeOwnersFields
    className?: string
}

export const EntityCodeOwners: React.FunctionComponent<Props> = ({ entity: { codeOwners }, className }) => (
    <PersonListRow
        title="Code owners"
        listTag="ol"
        items={
            codeOwners
                ? codeOwners.map(codeOwner => ({
                      person: codeOwner.node,
                      text:
                          codeOwner.fileProportion >= 0.01 ? `${(codeOwner.fileProportion * 100).toFixed(0)}%` : '<1%',
                      textTooltip: `Owns ${codeOwner.fileCount} ${pluralize('line', codeOwner.fileCount)}`,
                  }))
                : []
        }
        className={className}
    />
)
