import React, { useState, useEffect } from 'react'

import {
    Button,
    FormControl,
    FormGroup,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Stack,
    Typography,
    Checkbox,
    FormControlLabel,
} from '@mui/material'

import { changeStage } from './state'

export const Install: React.FC = () => {
    type installState = 'select-version' | 'select-db-type'
    const [installState, setInstallState] = useState<installState>('select-version')

    const [versions, setVersions] = useState<string[]>([])
    const [selectedVersion, setSelectedVersion] = useState<string>('')

    type dbType = 'internal' | 'external'
    const [dbType, setDbType] = useState<dbType>('internal')

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
                    const publicVersions = data
                        .filter(item => item.public)
                        .filter(item => !item.is_development)
                        .map(item => item.version)
                    setVersions(publicVersions)
                    setSelectedVersion(publicVersions[0]) // Set the first version as default
                }
            } catch (error) {
                console.error('Failed to fetch versions:', error)
            }
        }

        fetchVersions()
    }, [])

    const next = () => {
        if (selectedVersion === '') {
            alert('Please select a version')
            return
        }
        setInstallState('select-db-type')
    }

    const back = () => {
        setInstallState('select-version')
    }

    const install = () => {
        changeStage({ action: 'installing', data: selectedVersion })
    }

    return (
        // Render a version selection box followed by a database configuration screen, then an install prompt
        <div className="install">
            <Typography variant="h5">Let's get started...</Typography>
            <Paper elevation={3} sx={{ p: 4 }}>
                {installState === 'select-version' ? (
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
                        <Button variant="contained" sx={{ width: 200 }} onClick={next}>
                            Next
                        </Button>
                    </Stack>
                ) : installState === 'select-db-type' ? (
                    <Stack direction="column" spacing={2}>
                        <FormGroup row>
                            <FormControlLabel control={<Checkbox />} label="Internal DBs" />
                            <FormControlLabel disabled control={<Checkbox />} label="External DBs" />
                        </FormGroup>
                        <Stack direction="row" spacing={2}>
                            <Button variant="contained" sx={{ width: 200 }} onClick={back}>
                                Back
                            </Button>
                            <Button variant="contained" sx={{ width: 200 }} onClick={install}>
                                Install
                            </Button>
                        </Stack>
                    </Stack>
                ) : null}
            </Paper>
        </div>
    )
}
