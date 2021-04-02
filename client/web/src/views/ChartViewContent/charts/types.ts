import { MouseEvent } from 'react';

export interface DatumClickEvent {
    originEvent: MouseEvent<unknown>,
    link?: string
}

export type onDatumClick = (event: DatumClickEvent) => void;
