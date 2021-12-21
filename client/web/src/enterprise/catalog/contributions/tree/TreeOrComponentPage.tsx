import React from 'react'
import { Container } from '@sourcegraph/wildcard'

import { TreePage } from '../../../../repo/tree/TreePage'
import treePageStyles from '../../../../repo/tree/TreePage.module.scss'

interface Props extends React.ComponentPropsWithoutRef<typeof TreePage> {}

export const TreeOrComponentPage: React.FunctionComponent<Props> = () => (
    <Container className={treePageStyles.container}>hello world</Container>
)
