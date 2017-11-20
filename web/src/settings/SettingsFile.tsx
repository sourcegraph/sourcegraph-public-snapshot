import * as React from 'react'

export interface Props {
    children: string
}

export const SettingsFile = ({ children }: Props) => (
    <div key={0} className="settings-file__contents" dangerouslySetInnerHTML={{ __html: children }} />
)
