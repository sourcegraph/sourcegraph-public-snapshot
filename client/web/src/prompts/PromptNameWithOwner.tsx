import { type FunctionComponent } from 'react'

import { type PromptFields } from '../graphql-operations'

export const PromptNameWithOwner: FunctionComponent<{ prompt: Pick<PromptFields, 'name' | 'owner'> }> = ({
    prompt: {
        name,
        owner: { namespaceName },
    },
}) => (
    <>
        {namespaceName}/<strong>{name}</strong>
    </>
)
