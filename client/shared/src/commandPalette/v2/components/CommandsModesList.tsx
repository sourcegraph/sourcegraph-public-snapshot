import classNames from 'classnames'
import React from 'react'

interface CommandModesListProps {
    hasActiveTextDocument: boolean

    hasWorkspaceRoot: boolean
}

export const CommandsModesList: React.FC<CommandModesListProps> = ({ hasActiveTextDocument, hasWorkspaceRoot }) => (
    <ul>
        <li>
            Command <kbd>{'>'}</kbd>
        </li>
        <li className={classNames(!hasActiveTextDocument && 'text-muted')}>
            Fuzzy <kbd>$</kbd>
        </li>
        <li className={classNames(!hasActiveTextDocument && 'text-muted')}>
            Jump to line <kbd>:</kbd>
        </li>
        <li className={classNames(!hasActiveTextDocument && 'text-muted')}>
            Jump to symbol <kbd>@</kbd>
        </li>
        <li>
            Recent searches <kbd>#</kbd>
        </li>
    </ul>
)
