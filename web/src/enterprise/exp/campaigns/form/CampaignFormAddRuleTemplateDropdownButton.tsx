import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useState } from 'react'
import { CampaignFormControl } from './CampaignForm'
import { ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'
import { RULE_TEMPLATES, RuleTemplate } from './templates'
import { CampaignTemplateRow } from './CampaignTemplateRow'

interface Props extends CampaignFormControl {
    className?: string
}

export const CampaignFormAddRuleTemplateDropdownButton: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
    className = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        (template: RuleTemplate) => {
            const rules = value.rules || []
            onChange({
                rules: [
                    ...rules,
                    { name: template.title, definition: '{}', template: { template: template.id, context: {} } },
                ],
            })
        },
        [onChange, value.rules]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle color="" className={className} disabled={disabled}>
                <PlusIcon className="icon-inline" /> Add rule
            </DropdownToggle>
            <DropdownMenu style={{ maxWidth: '75vw', width: '30rem' }}>
                {RULE_TEMPLATES.filter(template => !template.isEmpty).map(template => (
                    <DropdownItem
                        key={template.id}
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => onSelect(template)}
                        className="d-flex align-items-start"
                        style={{ wordWrap: 'normal', whiteSpace: 'normal' }}
                    >
                        <CampaignTemplateRow template={template} />
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
