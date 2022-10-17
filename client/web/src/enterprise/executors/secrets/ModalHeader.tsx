import React from 'react'

import { H3 } from '@sourcegraph/wildcard'

export interface ModalHeaderProps {
    id: string
    secretKey: string
}

export const ModalHeader: React.FunctionComponent<React.PropsWithChildren<ModalHeaderProps>> = ({ id, secretKey }) => (
    <>
        <H3 id={id}>Executor secret: {secretKey}</H3>
    </>
)
