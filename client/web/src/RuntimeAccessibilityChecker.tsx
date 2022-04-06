import React, { useEffect, useState } from 'react'

import { debounce } from 'lodash'

import { runtimeAccessibilityAudit } from '@sourcegraph/shared/src/accessibility'
import { RuleViolation } from '@sourcegraph/shared/src/accessibility/formatAxeViolations'
import {
    Popover,
    PopoverTrigger,
    PopoverContent,
    Button,
    Position,
    MenuHeader,
    MenuDivider,
    PopoverOpenEvent,
    Modal,
} from '@sourcegraph/wildcard'

import styles from './RuntimeAccessibilityChecker.module.scss'

const runAudit = debounce(runtimeAccessibilityAudit, 500)

interface SummaryPopoverProps {
    isOpen: boolean
    handleOpenChange: (event: PopoverOpenEvent) => void
    handleMoreDetails: (details: ViolationDetails[]) => void
    violations: RuleViolation[]
}

const SummaryPopover: React.FunctionComponent<SummaryPopoverProps> = ({
    isOpen,
    handleOpenChange,
    handleMoreDetails,
    violations,
}) => {
    const [highlightedElements, setHighlightedElements] = useState<string[] | null>(null)

    // get total violations

    const title = `${violations.length} accessibility violation${violations.length === 1 ? '' : 's'}`

    useEffect(() => {
        const styleId = 'accessibility-highlighted-element'
        if (highlightedElements) {
            const style = document.createElement('style')
            const styles = highlightedElements
                .map(
                    target => `
                    ${target} {
                        border: 1px solid red;
                    }
                `
                )
                .join('\n')

            style.innerHTML = styles
            style.id = styleId
            document.body.append(style)
        }

        return () => {
            const style = document.querySelector(styleId)
            if (style) {
                style.remove()
            }
        }
    }, [highlightedElements])

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} variant="danger" className={styles.triggerButton}>
                {title}
            </PopoverTrigger>
            <PopoverContent position={Position.topStart} className={styles.content}>
                <MenuHeader>
                    <strong>{title}:</strong>
                </MenuHeader>
                <MenuDivider />
                {violations.map(({ details, error }, index) => (
                    <div key={index}>
                        <MenuHeader>{error}</MenuHeader>
                        <div className="px-2 mb-2">
                            <Button
                                className="mr-2"
                                variant="primary"
                                size="sm"
                                onClick={() => setHighlightedElements(details.flatMap(detail => detail.targets))}
                            >
                                Highlight elements
                            </Button>
                            <Button variant="secondary" size="sm" onClick={() => handleMoreDetails(details)}>
                                More details
                            </Button>
                        </div>
                    </div>
                ))}
            </PopoverContent>
        </Popover>
    )
}

interface ViolationDetails {
    summary: string
    targets: string[]
}

interface DetailsModalProps {
    details: ViolationDetails[]
    onDismiss: () => void
}

const DetailsModal: React.FunctionComponent<DetailsModalProps> = ({ details, onDismiss }) => {
    console.log('modal')

    return (
        <Modal onDismiss={onDismiss} aria-labelledby="">
            <h1>Violations</h1>
            {details.map(({ summary, targets }) => (
                <>
                    <pre>Elements: {targets.join(', ')}</pre>
                    <pre>{summary}</pre>
                </>
            ))}
        </Modal>
    )
}

export const AccessibilityRuntimeChecker: React.FunctionComponent = () => {
    const [popoverOpen, setPopoverOpen] = useState(false)
    const [modalDetails, setModalDetails] = useState<ViolationDetails>()
    const [violations, setViolations] = useState<RuleViolation[]>([])

    const onMutation: MutationCallback = async () => {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        const violations = await runAudit()
        setViolations(violations || [])
    }
    useEffect(() => {
        const rootElement = document.querySelector('#root')

        if (!rootElement) {
            throw new Error('Observer was called before React rendered')
        }

        const config = {
            attributes: true,
            childList: true,
            subtree: true,
        }
        const observer = new MutationObserver(onMutation)
        observer.observe(rootElement, config)

        return () => {
            observer.disconnect()
        }
    }, [])

    return (
        <>
            <SummaryPopover
                isOpen={popoverOpen}
                handleOpenChange={event => setPopoverOpen(event.isOpen)}
                handleMoreDetails={setModalDetails}
                violations={violations}
            />
            {modalDetails && <DetailsModal details={modalDetails} onDismiss={() => setModalDetails(undefined)} />}
        </>
    )
}
