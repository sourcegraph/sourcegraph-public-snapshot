import { H1 } from '@sourcegraph/wildcard'
import React from 'react'
import { ConsoleUserData } from '../../model'
import { InstanceList } from './InstanceList'

export const InstanceListPage: React.FunctionComponent<{ data: ConsoleUserData }> = ({ data }) => (
    <>
        <H1>Sourcegraph Cloud instances</H1>
        <InstanceList instances={data.instances} />
    </>
)
