import { useCallback, useEffect, useState } from 'react'

import { useSearchParams } from 'react-router-dom'

type UseInviteParamsHook = () => {
    inviteParams: { teamId: string; inviteId: string } | undefined
    clearInviteParams: () => void
}
export const useInviteParams: UseInviteParamsHook = () => {
    const [searchParams, setSearchParams] = useSearchParams()
    const [inviteParams, setInviteParams] = useState<{ teamId: string; inviteId: string }>()

    useEffect(() => {
        setInviteParams(s => {
            if (s) {
                return s
            }

            const teamId = searchParams.get('teamID')
            const inviteId = searchParams.get('inviteID')

            if (teamId && inviteId) {
                return { teamId, inviteId }
            }

            return undefined
        })
    }, [searchParams])

    const clearInviteParams = useCallback(
        () =>
            setSearchParams(params => {
                params.delete('teamID')
                params.delete('inviteID')
                return params
            }),
        [setSearchParams]
    )

    return { inviteParams, clearInviteParams }
}
