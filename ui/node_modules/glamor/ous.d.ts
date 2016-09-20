import { StyleAttribute } from './index';

type FractionalWidth = 'half' | 'oneThird' | 'twoThird';
type FractionalOffset = FractionalWidth;

export const container: StyleAttribute;
export const row: StyleAttribute;
export const columns: (columns: number | FractionalWidth, offset?: number | FractionalOffset) => StyleAttribute;
export const half: (offset?: number | FractionalOffset) => StyleAttribute;
export const oneThird: (offset?: number | FractionalOffset) => StyleAttribute;
export const twoThirds: (offset?: number | FractionalOffset) => StyleAttribute;
export const button: StyleAttribute;
export const primary: StyleAttribute;
export const labelBody: StyleAttribute;
export const base: StyleAttribute;
export const fullWidth: StyleAttribute;
export const maxFullWidth: StyleAttribute;
export const pullRight: StyleAttribute;
export const pullLeft: StyleAttribute;
export const clearfix: StyleAttribute;
