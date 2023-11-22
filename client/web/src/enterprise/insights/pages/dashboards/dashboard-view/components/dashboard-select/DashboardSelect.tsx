import { type ButtonHTMLAttributes, type ChangeEvent, type FC, forwardRef, useMemo, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import {
    Button,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionGroup,
    ComboboxOptionText,
    createRectangle,
    Flipping,
    Popover,
    PopoverContent,
    PopoverTrigger,
    usePopoverContext,
    Strategy,
    Badge,
    type PopoverOpenEvent,
    Link,
    Text,
    H3,
} from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../components'
import {
    type CustomInsightDashboard,
    isGlobalDashboard,
    isPersonalDashboard,
    isVirtualDashboard,
} from '../../../../../core'
import { useUiFeatures } from '../../../../../hooks'

import { getDashboardOwnerName, getDashboardOrganizationsGroups } from './helpers'

import styles from './DashboardSelect.module.scss'

const POPOVER_PADDING = createRectangle(0, 0, 5, 5)

export interface DashboardSelectProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onSelect'> {
    dashboard: CustomInsightDashboard | undefined
    dashboards: CustomInsightDashboard[]
    onSelect: (dashboard: CustomInsightDashboard) => void
}

export const DashboardSelect: FC<DashboardSelectProps> = props => {
    const { dashboard, dashboards, onSelect, ...attributes } = props

    const [isOpen, setOpen] = useState(false)

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        setOpen(event.isOpen)
    }

    const handleSelect = (dashboard: CustomInsightDashboard): void => {
        setOpen(false)
        onSelect(dashboard)
    }

    const hasDashboards = dashboards.length > 0
    const dashboardTitle = hasDashboards ? dashboard?.title : 'No dashboards to select'

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger
                {...attributes}
                as={DashboardSelectButton}
                title={dashboardTitle}
                badge={getDashboardOwnerName(dashboard)}
                disabled={!hasDashboards}
            />

            <PopoverContent
                targetPadding={POPOVER_PADDING}
                flipping={Flipping.opposite}
                strategy={Strategy.Absolute}
                className={styles.popover}
            >
                <DashboardSelectContent dashboard={dashboard} dashboards={dashboards} onSelect={handleSelect} />
            </PopoverContent>
        </Popover>
    )
}

interface DashboardSelectButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    title: string | undefined
    badge: string | undefined
}

const DashboardSelectButton = forwardRef<HTMLButtonElement, DashboardSelectButtonProps>((props, ref) => {
    const { title, badge, className, ...attributes } = props
    const { isOpen } = usePopoverContext()

    const Icon = isOpen ? ChevronUpIcon : ChevronDownIcon

    return (
        <Button
            {...attributes}
            ref={ref}
            variant="secondary"
            outline={true}
            aria-label={`Choose a dashboard, ${title}`}
            className={classNames(className, styles.triggerButton)}
        >
            <span className={styles.triggerButtonText}>
                <TruncatedText title={title}>{title ?? 'Unknown dashboard'}</TruncatedText>
                {badge && <InsightBadge value={badge} />}
            </span>

            <Icon className={styles.triggerButtonIcon} />
        </Button>
    )
})

interface DashboardSelectContentProps {
    dashboard: CustomInsightDashboard | undefined
    dashboards: CustomInsightDashboard[]
    onSelect: (dashboard: CustomInsightDashboard) => void
}

const DashboardSelectContent: FC<DashboardSelectContentProps> = props => {
    const { dashboard: currentDashboard, dashboards, onSelect } = props

    const { licensed } = useUiFeatures()
    const [state, setState] = useState(() => ({
        value: currentDashboard && isVirtualDashboard(currentDashboard) ? '' : currentDashboard?.title ?? '',
        changed: false,
    }))

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const inputValue = event.target.value

        setState({ value: inputValue, changed: true })
    }

    const handleSelect = (value: string): void => {
        const dashboard = dashboards.find(dashboard => dashboard.title === value)

        if (dashboard) {
            onSelect(dashboard)
        }
    }

    const filteredDashboards = useMemo(() => {
        if (state.changed) {
            return dashboards.filter(({ title }) => title.toLowerCase().includes(state.value.trim().toLowerCase()))
        }

        return dashboards
    }, [dashboards, state])

    const organizationGroups = useMemo(() => getDashboardOrganizationsGroups(filteredDashboards), [filteredDashboards])

    return (
        <Combobox
            openOnFocus={true}
            className={classNames(styles.combobox, { [styles.comboboxUnchanged]: !state.changed })}
            onSelect={handleSelect}
        >
            <ComboboxInput
                value={state.value}
                variant="small"
                autoFocus={true}
                spellCheck={false}
                placeholder="Find a dashboard..."
                aria-label="Find a dashboard"
                inputClassName={styles.comboboxInput}
                className={styles.comboboxInputContainer}
                onChange={handleInputChange}
            />

            <ComboboxList persistSelection={true} className={styles.comboboxList}>
                {filteredDashboards.some(isPersonalDashboard) && (
                    <ComboboxOptionGroup heading="Private" className={styles.comboboxOptionGroup}>
                        {filteredDashboards.filter(isPersonalDashboard).map(dashboard => (
                            <DashboardOption
                                key={dashboard.id}
                                name={dashboard.title}
                                selected={dashboard.id === currentDashboard?.id}
                                badgeText={getDashboardOwnerName(dashboard)}
                            />
                        ))}
                    </ComboboxOptionGroup>
                )}

                {filteredDashboards.some(isGlobalDashboard) && (
                    <ComboboxOptionGroup heading="Global" className={styles.comboboxOptionGroup}>
                        {filteredDashboards.filter(isGlobalDashboard).map(dashboard => (
                            <DashboardOption
                                key={dashboard.id}
                                name={dashboard.title}
                                selected={dashboard.id === currentDashboard?.id}
                                badgeText={getDashboardOwnerName(dashboard)}
                            />
                        ))}
                    </ComboboxOptionGroup>
                )}

                {organizationGroups.map(group => (
                    <ComboboxOptionGroup key={group.id} heading={group.name} className={styles.comboboxOptionGroup}>
                        {group.dashboards.map(dashboard => (
                            <DashboardOption
                                key={dashboard.id}
                                name={dashboard.title}
                                selected={dashboard.id === currentDashboard?.id}
                                badgeText={getDashboardOwnerName(dashboard)}
                            />
                        ))}
                    </ComboboxOptionGroup>
                ))}

                {filteredDashboards.length === 0 && (
                    <div className={styles.noResultsFound}>
                        <Text as="span" className="text-muted">
                            No dashboards found.
                        </Text>
                        <Button as={Link} variant="link" to="/insights/add-dashboard">
                            Create a dashboard
                        </Button>
                    </div>
                )}

                {!licensed && (
                    <div>
                        <hr />

                        <div className={classNames(styles.limitedAccess)}>
                            <H3>Limited access</H3>
                            <Text>Unlock for unlimited custom dashboards.</Text>
                        </div>
                    </div>
                )}
            </ComboboxList>
        </Combobox>
    )
}

interface DashboardOptionProps {
    name: string
    selected: boolean
    badgeText?: string
}

const DashboardOption: FC<DashboardOptionProps> = props => {
    const { name, selected, badgeText } = props

    return (
        <ComboboxOption
            value={name}
            selected={selected}
            data-value={name}
            className={classNames(styles.comboboxOption, { [styles.comboboxOptionSelected]: selected })}
        >
            <TruncatedText title={name}>
                <ComboboxOptionText />
            </TruncatedText>
            {badgeText && <InsightBadge value={badgeText} />}
        </ComboboxOption>
    )
}

interface BadgeProps {
    value: string
    className?: string
}

const InsightBadge: FC<BadgeProps> = props => {
    const { value, className } = props

    return (
        <TruncatedText as={Badge} title={value} variant="secondary" className={classNames(styles.badge, className)}>
            {value}
        </TruncatedText>
    )
}
