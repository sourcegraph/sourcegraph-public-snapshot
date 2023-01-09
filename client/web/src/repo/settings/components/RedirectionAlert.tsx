import { FC, useEffect, useState } from 'react'

import { useHistory } from 'react-router'

import { Alert } from '@sourcegraph/wildcard'

interface Props {
    to: string
    messagePrefix: string
    className?: string
}

/**
 * The repository settings options page.
 */
export const RedirectionAlert: FC<Props> = ({ to, className, messagePrefix }) => {
    const [ttl, setTtl] = useState(3)
    const history = useHistory()

    useEffect(() => {
        const interval = setInterval(() => setTtl(ttl => ttl - 1), 700)

        return () => clearInterval(interval)
    }, [])

    useEffect(() => {
        if (ttl === 0) {
            history.push(to)
        }
    }, [ttl, history, to])

    return (
        <Alert className={className} variant="success">
            {messagePrefix} You will be redirected in {ttl}...
        </Alert>
    )
}
