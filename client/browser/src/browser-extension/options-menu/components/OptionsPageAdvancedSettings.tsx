import React, { useEffect, useContext, useState } from 'react'

import { OptionsPageContext } from '../OptionsPage.context'

import { CodeTextArea } from './CodeTextArea'
import { InfoText } from './InfoText'

const PLACEHOLDER = ['https://github.com/org/repo', 'github.com/org/*'].join('\n')

interface CheckboxProps {
    value: boolean
    onChange: (value: boolean) => void
    dataTestId?: string
}
const Checkbox: React.FC<CheckboxProps> = ({ value, children, onChange, dataTestId }) => (
    <div className="form-check">
        <label className="form-check-label" data-testid={dataTestId}>
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
            <Checkbox
                value={isBlocklistEnabled}
                onChange={setIsBlocklistEnabled}
                dataTestId="test-cloud-blocklist-toggle"
            >
                Sourcegraph cloud repository blocklist
            </Checkbox>
            {isBlocklistEnabled && (
                <>
                    {blocklistContent === blocklist?.content && <span data-testid="blocklist-is-saved" />}
                    <InfoText className="m-2">
                        We won’t make any requests to Sourcegraph cloud for the URLs that are listed here.
                    </InfoText>
                    <CodeTextArea
                        dataTestId="test-cloud-blocklist-textarea"
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
