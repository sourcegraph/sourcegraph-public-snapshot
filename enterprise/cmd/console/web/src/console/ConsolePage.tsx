import { H1 } from '@sourcegraph/wildcard'
import React from 'react'
import { InstanceList } from '../instances/InstanceList'
import { ConsoleUserData } from '../model'
import { ConsoleLayout } from './ConsoleLayout'

export const ConsolePage: React.FunctionComponent<{ data: ConsoleUserData }> = ({ data }) => (
    <ConsoleLayout data={data}>
        <H1>Sourcegraph Cloud instances</H1>
        <InstanceList instances={data.instances} />
    </ConsoleLayout>
)
