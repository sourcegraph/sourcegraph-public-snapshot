export enum PopoverOpenEventReason {
    TriggerClick = 'TriggerClick',
    TriggerFocus = 'TriggerFocus',
    TriggerBlur = 'TriggerBlur',
    ClickOutside = 'ClickOutside',
    Esc = 'Esc',
}

export interface PopoverOpenEvent {
    isOpen: boolean
    reason: PopoverOpenEventReason
}
