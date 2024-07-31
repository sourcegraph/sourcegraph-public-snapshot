import React, { useEffect, useState } from 'react'

import { Button, CircularProgress, Stack, Typography } from '@mui/material'

import { changeStage } from './state.ts'

const WaitBeforeLaunchMs = 3 * 1000

export const WaitForAdmin: React.FC = () => {
    const [launching, setLaunching] = useState<boolean>(false)

    useEffect(() => {
        if (launching) {
            const timer = setInterval(() => {
                changeStage({ action: 'refresh' })
            }, WaitBeforeLaunchMs)
            return () => clearInterval(timer)
        }
    }, [launching])

    return (
        <div className="wait-for-admin">
            <Typography variant="h4">Waiting For The Admin To Return</Typography>
            <div>
                <Typography sx={{ m: 2 }}>
                    The appliance is ready. We were waiting for you to set its security before opening it up.
                </Typography>
                <Typography sx={{ m: 2 }}>
                    Now that you're back, please press the button below to launch the Administration UI.
                </Typography>
            </div>
            <Button variant="contained" onClick={() => setLaunching(true)} disabled={launching}>
                Launch Admin UI
            </Button>
            {launching && (
                <Stack direction="row" spacing={2}>
                    <CircularProgress size={32} />
                    <Typography variant="h5">Launching Admin UI... Please wait...</Typography>
                </Stack>
            )}
        </div>
    )
}
