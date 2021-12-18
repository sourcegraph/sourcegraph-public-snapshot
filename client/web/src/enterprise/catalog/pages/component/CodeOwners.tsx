import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ComponentCodeOwnersFields } from '../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    entity: ComponentCodeOwnersFields
    className?: string
}

export const EntityCodeOwners: React.FunctionComponent<Props> = ({ entity: { codeOwners }, className }) => (
    <PersonList
        title="Code owners"
        listTag="ol"
        orientation="vertical"
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
