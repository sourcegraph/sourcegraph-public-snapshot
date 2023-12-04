import type { LicenseKeyInfo } from './types'

export const DEFAULT_FORMAT_OPTIONS = {
    formattingOptions: {
        eol: '\n',
        insertSpaces: true,
        tabSize: 2,
    },
}

export const DEBOUNCE_DELAY_MS = 500
export const MIN_INPUT_LENGTH = 10

export const DEFAULT_LICENSE_KEY_INFO: LicenseKeyInfo = {
    title: 'Current License',
    type: 'Free',
    description: '1-user license, valid indefinitely',
    logo: (
        <div
            // eslint-disable-next-line react/forbid-dom-props
            style={{
                width: '50px',
                height: '50px',
                alignSelf: 'center',
                justifySelf: 'center',
                borderRadius: '50%',
                border: '1px solid #CAD2E2',
            }}
        />
    ),
}
