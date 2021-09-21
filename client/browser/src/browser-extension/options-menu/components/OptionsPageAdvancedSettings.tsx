import React, { useEffect, useContext, useState } from 'react'

import { OptionsPageContext } from '../OptionsPage.context'

import { CodeTextArea } from './CodeTextArea'
import { InfoText } from './InfoText'

const PLACEHOLDER = ['https://github.com/org/repo', 'github.com/org/*'].join('\n')

const Checkbox: React.FC<{ value: boolean; onChange: (value: boolean) => void }> = ({ value, children, onChange }) => (
    <div className="form-check">
        <label className="form-check-label">
            <input
                onChange={event => onChange(event.target.checked)}
                className="form-check-input"
                type="checkbox"
                checked={value}
            />{' '}
            {children}
        </label>
    </div>
)

export const OptionsPageAdvancedSettings: React.FC = () => {
    const { optionFlags, onChangeOptionFlag, blocklist, onBlocklistChange } = useContext(OptionsPageContext)
    const [isBlocklistEnabled, setIsBlocklistEnabled] = useState(!!blocklist?.enabled)
    const [blocklistContent, setBlocklistContent] = useState(blocklist?.content ?? '')

    useEffect(() => {
        onBlocklistChange(isBlocklistEnabled, blocklistContent)
    }, [isBlocklistEnabled, blocklistContent, onBlocklistChange])

    return (
        <section className="mt-3 mb-2">
            {optionFlags.map(({ label, key, value }) => (
                <Checkbox key={key} value={value} onChange={value => onChangeOptionFlag(key, value)}>
                    {label}
                </Checkbox>
            ))}
            <Checkbox value={isBlocklistEnabled} onChange={setIsBlocklistEnabled}>
                Sourcegraph cloud repository blocklist
            </Checkbox>
            {isBlocklistEnabled && (
                <>
                    <InfoText className="m-2">
                        We won’t make any requests to Sourcegraph cloud for the URLs that are listed here.
                    </InfoText>
                    <CodeTextArea
                        rows={4}
                        placeholder={PLACEHOLDER}
                        value={blocklistContent}
                        onChange={setBlocklistContent}
                    />
                </>
            )}
        </section>
    )
}
