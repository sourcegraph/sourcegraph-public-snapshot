import React from 'react'

/**
 * Describes a section in a {@link Sections} component.
 *
 * @template ID The type that includes all possible section IDs (typically a union of string constants).
 */
export interface Section<ID extends string> {
    id: ID
    label: React.ReactFragment

    /**
     * Whether the section is hidden.
     */
    hidden?: boolean
}

/**
 * Properties for the Sections components and its wrappers.
 *
 * @template ID The type that includes all possible section IDs (typically a union of string constants).
 * @template T The type that describes a section.
 */
export interface SectionsProps<ID extends string, T extends Section<ID>> {
    /** All sections. */
    sections: T[]

    /** The currently active section. */
    activeSection: ID | undefined

    /**
     * The component used to render a section's navbar item.
     */
    navbarItemComponent: React.ComponentType<{ section: T; className: string }>

    children: React.ReactFragment

    id?: string
    className?: string
    navbarItemClassName?: string

    /** Optional handler when a navbar item is selected */
    onSelectNavbarItem?: (section: ID) => void
}

/**
 * Properties for Sections components that provide their own state persistence.
 */
export interface SectionsWithPersistenceProps<ID extends string, T extends Section<ID>>
    extends Omit<SectionsProps<ID, T>, 'activeSection' | 'navbarItemComponent'> {}
