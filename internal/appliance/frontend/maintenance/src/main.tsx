import React from 'react'

import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'

import { Frame } from './Frame'
import { Home } from './Home'

import './index.css'

import { Install } from './Install'
import { Maintenance } from './Maintenance'
import { Progress } from './Progress'
import { ThemeProvider } from './Theme'
import { WaitForAdmin } from './WaitForAdmin'

import reportWebVitals from './reportWebVitals';

const router = createBrowserRouter([
    {
        path: '/',
        element: <Frame />,
        children: [
            {
                path: '',
                element: <Home />,
            },
            {
                path: 'install',
                element: <Install />,
            },
            {
                path: 'install/progress',
                element: <Progress action="install" />,
            },
            {
                path: 'install/wait-for-admin',
                element: <WaitForAdmin />,
            },
            {
                path: 'upgrade/progress',
                element: <Progress action="upgrade" />,
            },
            {
                path: 'maintenance',
                element: <Maintenance />,
            },
        ],
    },
])

ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
        <ThemeProvider>
            <RouterProvider router={router} />
        </ThemeProvider>
    </React.StrictMode>
)

reportWebVitals(console.log)
