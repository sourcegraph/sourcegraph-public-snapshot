import React, { useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

import { createTether, Flipping, Overlapping, Position, Tether } from '../tether';

import styles from './FloatingPanel.module.scss'

export interface FloatingPanelProps extends Omit<Tether, 'target' | 'element' | 'marker'> {
	/**
	 * Reference on target HTML element in the DOM.
	 * Renders nothing if target isn't specified.
	 */
	target: HTMLElement | null

	/**
	 * Enables tail element rendering and attaches it to
	 * floating panel.
	 */
	marker?: boolean;

	className?: string;
}

/**
 * React component that wraps up tether positioning logic and provide narrowed down
 * interface of setting to setup floating panel component.
 */
export const FloatingPanel: React.FunctionComponent<FloatingPanelProps> = props => {
	const {
		target,
		marker,
		position = Position.bottomLeft,
		overlapping = Overlapping.none,
		flipping = Flipping.opposite,
		pin = null,
		constrainToScrollParents = true,
		overflowToScrollParents = true,
		windowPadding,
		constraintPadding,
		constraint,
		className = '',
	} = props;

	const containerReference = useRef(document.createElement('div'));
	const [tooltipElement, setTooltipElement] = useState<HTMLDivElement>()
	const [tooltipTailElement, setTooltipTailElement] = useState<HTMLDivElement>()

	const setTooltipReference = (node: HTMLDivElement | null): void => {
		if (node) {
            setTooltipElement(node)
        }
	}

	const setTooltipTailReference = (node: HTMLDivElement | null): void => {
		if (node) {
            setTooltipTailElement(node)
        }
	}

	// Add a container element right after the body tag
	useLayoutEffect(() => {
        const element = containerReference.current

		document.body.append(element);

		return () => { element.remove(); }
	}, [containerReference]);

	useLayoutEffect(() => {
		if (!tooltipElement) {
			return
		}

		const { unsubscribe } = createTether({
			element: tooltipElement,
            marker: tooltipTailElement,
            target,
            constraint,
			pin,
			windowPadding,
			constraintPadding,
			position,
			overlapping,
			constrainToScrollParents,
			overflowToScrollParents,
			flipping,
		});

		return unsubscribe
	}, [
		target,
		tooltipElement,
		tooltipTailElement,
        constraint,
		windowPadding,
		constraintPadding,
		pin,
		position,
		overlapping,
		constrainToScrollParents,
		overflowToScrollParents,
		flipping,
	]);

	return createPortal(
		<>
			<div ref={setTooltipReference} className={`${styles.floatingPanel} ${className}`}>
				{props.children}
			</div>

			{marker && <div className={styles.tail} ref={setTooltipTailReference}/>}
		</>,
		containerReference.current,
	);
}
