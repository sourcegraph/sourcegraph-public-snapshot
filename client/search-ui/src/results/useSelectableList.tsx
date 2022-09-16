import { useCallback, useEffect, useLayoutEffect } from 'react'
import { useHistory } from 'react-router'

import styles from './useSelectableList.module.scss'

export function useSelectableList(): void {
    const history = useHistory()
    const key = useCallback(() => history.createHref(history.location), [])
    const highlightedHref: () => string | null = useCallback(() => sessionStorage.getItem(key()), [])
    // const initialIndex = sessionStorage.getItem(history.createHref(history.location))
    // const [highlightedIndex, setHighlightedIndex] = useState(initialIndex === null ? 0 : Number(initialIndex))

    const selectableEntries = useCallback(() => document.querySelectorAll('[data-selectable-href]'), [])

    const highlightNext = (up: boolean): void => {
        const highlighted = highlightedHref()
        if (!highlighted) {
            return
        }
        let entries: Iterable<Element> = selectableEntries()
        if (up) {
            entries = [...entries].reverse()
        }
        let previousElement: Element | undefined
        for (const element of entries) {
            const href = element.getAttribute('data-selectable-href')
            if (href === null) {
                continue
            }
            if (previousElement) {
                element.classList.add(styles.highlighted)
                previousElement.classList.remove(styles.highlighted)
                sessionStorage.setItem(key(), href)
                element.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'center' })
                break
            }
            if (href === highlighted) {
                previousElement = element
            }
        }
    }

    const onKeyDown = useCallback((event: KeyboardEvent) => {
        const target = event.target as HTMLElement
        const shouldContiue = target.nodeName === 'BODY' || target.classList.contains('selectable-list-compatible')
        if (!shouldContiue) {
            // Only update higlighted item when the focus is on the body
            // element or an element that is compatible with selectable
            // list.
            return
        }

        switch (event.key) {
            case 'k':
            case 'j':
            case 'ArrowUp':
            case 'ArrowDown':
                // TODO: reject k and j when there are modifiers
                event.preventDefault()
                event.stopPropagation()
                highlightNext(event.key === 'ArrowUp' || event.key === 'k')
                break
            case 'Enter':
                const highlighted = highlightedHref()
                for (const element of selectableEntries()) {
                    const href = element.getAttribute('data-selectable-href')
                    if (href !== null && href === highlighted) {
                        history.push(href)
                    }
                }
                break
            default:
                if (/[a-zA-Z]/.test(event.key) && !event.altKey && !event.ctrlKey && !event.shiftKey) {
                    const toast = document.createElement('div')
                    toast.innerHTML = 'Press / to jump to the search box'
                    toast.style.border = '1px solid black'
                    toast.style.borderRadius = '10px'
                    toast.style.width = 'auto'
                    toast.style.backgroundColor = 'grey'
                    toast.style.fontSize = '1em'
                    toast.style.color = 'white'
                    toast.style.padding = '10px'
                    toast.style.position = 'absolute'
                    toast.style.left = '10px'
                    toast.style.bottom = '10px'
                    console.log(toast)
                    document.body.appendChild(toast)
                    setTimeout(() => {
                        toast.remove()
                    }, 3000)
                }
        }
    }, [])

    useLayoutEffect(() => {
        if (document.querySelector('.' + styles.highlighted)) {
            // Do nothing, there is already a highlighted element.
            return
        }
        // TODO : figure out how to re-run this effect after user collapses a highlighted search result.
        const highlighted = highlightedHref()
        let firstElement: [Element, string] | undefined
        let elementToHighlight: [Element, string] | undefined
        for (const element of selectableEntries()) {
            const href = element.getAttribute('data-selectable-href')
            if (!href) {
                continue
            }
            if (highlighted === href) {
                elementToHighlight = [element, href]
                break
            }
            if (!firstElement) {
                firstElement = [element, href]
            }
        }
        if (!elementToHighlight) {
            elementToHighlight = firstElement
        }
        if (elementToHighlight) {
            sessionStorage.setItem(key(), elementToHighlight[1])
            elementToHighlight[0].classList.add(styles.highlighted)
            elementToHighlight[0].scrollIntoView({ behavior: 'auto', inline: 'center', block: 'center' })
        }
    })

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown, { capture: true })
        return () => document.removeEventListener('keydown', onKeyDown, { capture: true })
    })
}
