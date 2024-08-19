import React, { useState, useEffect } from 'react'

import {
    Button,
    FormControl,
    RadioGroup,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Stack,
    Typography,
    Radio,
    FormControlLabel,
    FormGroup,
    FormLabel,
    FormHelperText,
    Box,
    TextField,
    Tab,
    Tabs,
} from '@mui/material'

import { call } from './api'
import { changeStage } from './state'

export const Install: React.FC = () => {
    type installState = 'select-version' | 'select-db-type'
    const [installState, setInstallState] = useState<installState>('select-version')

    const [versions, setVersions] = useState<string[]>([])
    const [selectedVersion, setSelectedVersion] = useState<string>('')

    type dbType = 'built-in' | 'external'
    const [dbType, setDbType] = useState<dbType>('built-in')

    type dbTab = 'pgsql' | 'codeintel' | 'codeinsights'
    const [dbTab, setDbTab] = useState<dbTab>('pgsql')

    const handleDbTabChange = (event: React.SyntheticEvent, newValue: dbTab) => {
        setDbTab(newValue)
    }

    useEffect(() => {
        const fetchVersions = async () => {
            try {
                const response = await call('/api/v1/releases/sourcegraph')
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

                // Very basic fallback for when release registry is down:
                // hardcode a particular version of Sourcegraph, which is the
                // latest at the time of writing.
                // This could be replaced with a fallback to a release registry
                // response fixture that appliance-frontend has access to on the
                // filesystem. In Kubernetes, this could be derived from a
                // ConfigMap, with the files being distributed to airgap users
                // out-of-band.
                const publicVersions = ['v5.5.2463']
                setVersions(publicVersions)
                setSelectedVersion(publicVersions[0])
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

    const handleDbSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
        setDbType(event.target.value as dbType)
    }

    return (
        // Render a version selection box followed by a database configuration screen, then an install prompt
        <div className="install">
            <Typography variant="h5">Setup Sourcegraph</Typography>
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
                            <Typography variant="caption">Proceed to database configuration.</Typography>
                        </div>
                        <Button variant="contained" sx={{ width: 200 }} onClick={next}>
                            Next
                        </Button>
                    </Stack>
                ) : installState === 'select-db-type' ? (
                    <Stack direction="column" spacing={2} alignItems={'center'}>
                        <FormControl>
                            <FormLabel>Configure Sourcegraph Databases</FormLabel>
                            <FormGroup>
                                <RadioGroup value={dbType} onChange={handleDbSelect} defaultValue="built-in">
                                    <FormControlLabel value="built-in" control={<Radio />} label="built-in DBs" />
                                    <FormHelperText id="my-helper-text" fontSize="small">
                                        Selecting built-in dbs, configures sourcegraph to use built in databases.
                                        Provisioned and controlled directly by appliance.{' '}
                                    </FormHelperText>
                                    <FormControlLabel
                                        value="external"
                                        control={<Radio />}
                                        label="External DBs (not yet supported)"
                                    />
                                </RadioGroup>
                            </FormGroup>
                        </FormControl>
                        {dbType === 'external' ? (
                            <Box sx={{ width: '80%' }} alignContent={'center'}>
                                <Box
                                    alignContent={'center'}
                                    sx={{ paddingBottom: 2.5, borderBottom: 1, borderColor: 'divider' }}
                                >
                                    <Tabs value={dbTab} onChange={handleDbTabChange}>
                                        <Tab label="Pgsql" disabled />
                                        <Tab label="Codeintel-db" disabled />
                                        <Tab label="Codeinsights-db" disabled />
                                    </Tabs>
                                </Box>
                                <FormGroup>
                                    <Stack spacing={2}>
                                        <TextField disabled label="Port" defaultValue="5432" />
                                        <TextField disabled label="User" defaultValue="sg" />
                                        <TextField disabled label="Password" defaultValue="sg" />
                                        <TextField disabled label="Database" defaultValue="sg" />
                                        <TextField disabled label="SSL Mode" defaultValue="disable" />
                                    </Stack>
                                </FormGroup>
                            </Box>
                        ) : null}
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
