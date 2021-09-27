import React from 'react'

export const CommandsModesList: React.FC = () => (
    <ul>
        <li>
            Command <kbd>{'>'}</kbd>
        </li>
        <li>
            Fuzzy <kbd>$</kbd>
        </li>
        <li>
            Jump to line <kbd>:</kbd>
        </li>
        <li>
            Jump to symbol <kbd>@</kbd>
        </li>
        <li>
            Recent searches <kbd>#</kbd>
        </li>
    </ul>
)
