import React, {useState} from 'react'
import {Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, Typography} from '@mui/material'
import {changeStage} from './state'

export const Install: React.FC = () => {
    const [version, setVersion] = useState<string>('5.5.0')

    const install = () => {
        changeStage({ action: 'installing', data: version })
    }

    return (
        <div className="install">
            <Typography variant="h5">Let's get started...</Typography>
            <Paper elevation={3} sx={{p: 4}}>
                <Stack direction="column" spacing={2} sx={{alignItems: 'center'}}>
                    <FormControl sx={{minWidth: 200}}>
                        <InputLabel id="demo-simple-select-label">Version</InputLabel>
                        <Select
                            value={version}
                            label="Age"
                            onChange={e => setVersion(e.target.value)}
                            sx={{width: 200}}
                        >
                            <MenuItem value={'5.5.0'}>5.5.0</MenuItem>
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
