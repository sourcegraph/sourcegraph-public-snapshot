import * as React from 'react'

export interface OptionsActionButtonProps {
    className?: string
    onClick: (event: React.MouseEvent<HTMLButtonElement>) => void
}

export const OptionsActionButton: React.SFC<OptionsActionButtonProps> = ({ children, className, onClick }) => (
    <button className={`options-action-button ${className || ''}`} onClick={onClick}>
        {children}
    </button>
)
