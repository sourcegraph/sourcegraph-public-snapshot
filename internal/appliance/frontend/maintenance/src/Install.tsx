import React, { useState, useEffect } from 'react'

import { Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, Typography } from '@mui/material'

import { changeStage } from './state'

export const Install: React.FC = () => {
    const [versions, setVersions] = useState<string[]>([])
    const [selectedVersion, setSelectedVersion] = useState<string>('')

    useEffect(() => {
        const fetchVersions = async () => {
            try {
                const response = await fetch('https://releaseregistry.sourcegraph.com/v1/releases/sourcegraph', {
                    headers: {
                        Authorization: `Bearer token`,
                        'Content-Type': 'application/json',
                    },
                    mode: 'cors',
                })
                const data = await response.json()
                setVersions(data)
                if (data.length > 0) {
                    const publicVersions = data.filter(item => item.public).map(item => item.version)
                    setVersions(publicVersions)
                    setSelectedVersion(publicVersions[0]) // Set the first version as default
                }
            } catch (error) {
                console.error('Failed to fetch versions:', error)
            }
        }

        fetchVersions()
    }, [])

    const install = () => {
        changeStage({ action: 'installing', data: selectedVersion })
    }

    return (
        <div className="install">
            <Typography variant="h5">Let's get started...</Typography>
            <Paper elevation={3} sx={{ p: 4 }}>
                <Stack direction="column" spacing={2} sx={{ alignItems: 'center' }}>
                    <FormControl sx={{ minWidth: 200 }}>
                        <InputLabel id="demo-simple-select-label">Version</InputLabel>
                        <Select
                            value={selectedVersion}
                            label="Version"
                            onChange={e => setSelectedVersion(e.target.value)}
                            sx={{ width: 200 }}
                        >
                            {versions.map(version => (
                                <MenuItem key={version} value={version}>
                                    {version}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                    <div className="message">
                        <Typography variant="caption">Press install to begin installation.</Typography>
                    </div>
                    <Button variant="contained" sx={{ width: 200 }} onClick={install}>
                        Install
                    </Button>
                </Stack>
            </Paper>
        </div>
    )
}
