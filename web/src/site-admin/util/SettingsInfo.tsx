import format from 'date-fns/format'
import * as React from 'react'

export const SettingsInfo: React.SFC<{ settings: GQL.ISettings; filename: string }> = props => (
    <span>
        Settings:{' '}
        <a
            href={encodeSettingsFile(props.settings.configuration.contents)}
            download={props.filename}
            target="_blank"
            title={props.settings.configuration.contents}
        >
            download JSON file
        </a>{' '}
        (saved on {format(props.settings.createdAt, 'YYYY-MM-DD')})
    </span>
)

function encodeSettingsFile(contents: string): string {
    return `data:application/json;charset=utf-8;base64,${btoa(contents)}`
}
