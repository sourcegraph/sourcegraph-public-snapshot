import React from 'react'
import { ConsoleUserData } from '../model'
import { ConsoleLayout } from './ConsoleLayout'

export const ConsolePage: React.FunctionComponent<{ data: ConsoleUserData }> = ({ data }) => (
    <ConsoleLayout data={data}>asdf</ConsoleLayout>
)
