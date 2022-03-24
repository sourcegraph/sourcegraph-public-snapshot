import React from 'react'

import AccountGroupIcon from 'mdi-react/AccountGroupIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CursorPointerIcon from 'mdi-react/CursorPointerIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

export const IconResolveIncidentsFaster: React.FunctionComponent = () => (
    <span className="mr-2">
        <svg width="24" height="53" viewBox="0 0 24 53" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M22.7697 1.8089C21.2815 0.101715 3.12525 0.0260916 1.33939 1.8089C-0.446466 3.59172 -0.446463 20.2188 1.33939 22.7487C3.12525 25.2786 20.9838 24.8805 22.7697 22.7487C24.5556 20.6169 24.2579 3.51609 22.7697 1.8089ZM8.48283 4.74074V20.0966L18.0074 12.4187L8.48283 4.74074Z"
                fill="currentColor"
            />
            <path
                d="M10.1007 32.0999C13.7007 32.0999 16.6007 34.9999 16.6007 38.5999C16.6007 40.1999 16.0007 41.6999 15.0007 42.7999L15.3007 43.0999H16.1007L21.1007 48.0999L19.6007 49.5999L14.6007 44.5999V43.7999L14.3007 43.4999C13.2007 44.4999 11.7007 45.0999 10.1007 45.0999C6.50071 45.0999 3.60071 42.1999 3.60071 38.5999C3.60071 34.9999 6.50071 32.0999 10.1007 32.0999ZM10.1007 34.0999C7.60071 34.0999 5.60071 36.0999 5.60071 38.5999C5.60071 41.0999 7.60071 43.0999 10.1007 43.0999C12.6007 43.0999 14.6007 41.0999 14.6007 38.5999C14.6007 36.0999 12.6007 34.0999 10.1007 34.0999Z"
                fill="currentColor"
            />
        </svg>
    </span>
)

export const IconPowerfulCodeNavigation: React.FunctionComponent = () => <CursorPointerIcon size="2rem" />
export const IconInstallIDEExtension: React.FunctionComponent = () => <PuzzleOutlineIcon size="2rem" />
export const IconCreateTeam: React.FunctionComponent = () => <AccountGroupIcon size="2rem" />
export const IconFindCodeReference: React.FunctionComponent = () => <MagnifyIcon size="2rem" />
export const IconAllDone: React.FunctionComponent = () => <CheckCircleIcon size="2rem" className="text-success" />
