import React, { FunctionComponent } from 'react'

import { Terminal } from '@sourcegraph/wildcard/src/components/Terminal'

import { useRepoCloningStatus } from './useRepoCloningStatus'

interface Props {
    userId: string
    pollInterval: number
}

export const CloningStatusTerminal: FunctionComponent<Props> = ({ userId, pollInterval = 2000 }) => {
    const { repos, loading, error, isDoneCloning } = useRepoCloningStatus({ userId, pollInterval })
    // parsed array for repo lines
    console.log(`Is fetching: ${loading}\nIs done cloning: ${isDoneCloning}\nIndividual repo lines =>`, repos)

    if (error) {
        console.log('CloningStatusTerminal =>', error)
    }

    // lines can be an empty array
    // isLoading is also passed - so we can render the terminal and display some "Loading..." message
    return <Terminal lines={repos} isLoading={loading} />
}
