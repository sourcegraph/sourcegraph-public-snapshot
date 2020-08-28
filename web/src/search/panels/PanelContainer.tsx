import * as React from 'react'

interface Props {
    title: string
}

export const PanelContainer: React.FunctionComponent<Props> = ({ title }) => (
    <>
        <div>{title}</div>
    </>
)
