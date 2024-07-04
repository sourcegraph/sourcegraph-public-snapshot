import styled from '@emotion/styled'
import { AppBar, Toolbar, Typography } from '@mui/material'
import { Outlet } from 'react-router-dom'

const Content = styled.div`
    margin: 1rem;
    margin-top: calc(64px + 1rem);
`

function App() {
  return (
    <>
        <AppBar>
            <Toolbar>
                <Typography variant="h5">Sourcegraph Appliance Administration</Typography>
            </Toolbar>
        </AppBar>
        <Content>
            <Outlet />
        </Content>
    </>
  )
}

export default App
