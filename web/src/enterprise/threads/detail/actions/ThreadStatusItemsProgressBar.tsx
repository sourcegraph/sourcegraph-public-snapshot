import React from 'react'

// tslint:disable: jsx-ban-props
export const ThreadStatusItemsProgressBar: React.FunctionComponent<{
    label?: string
    height?: string
    className?: string
}> = ({ label, height = '0.5rem', className = '' }) => (
    <div className={`progress ${className}`} style={{ height }}>
        <div
            className="progress-bar bg-success"
            role="progressbar"
            style={{ width: '50%' }}
            aria-valuenow={50}
            aria-valuemin={0}
            aria-valuemax={100}
            data-tooltip="Merged"
        >
            <strong>{label}</strong>
        </div>
        <div
            className="progress-bar bg-danger"
            role="progressbar"
            style={{ width: '20%' }}
            aria-valuenow={20}
            aria-valuemin={0}
            aria-valuemax={100}
            data-tooltip="Closed"
        />
        <div
            className="progress-bar"
            role="progressbar"
            style={{ width: '50%', backgroundColor: 'rgba(155,155,155,0.3)' }}
            aria-valuenow={50}
            aria-valuemin={0}
            aria-valuemax={100}
            data-tooltip="Open (not yet merged)"
        />
    </div>
)
