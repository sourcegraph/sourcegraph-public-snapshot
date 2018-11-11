import * as sourcegraph from 'sourcegraph';
import * as plain from '../../protocol/plainTypes';
import { Position } from '../types/position';
import { Range } from '../types/range';
/**
 * Converts from a plain object {@link plain.Position} to an instance of {@link Position}.
 *
 * @internal
 */
export declare function toPosition(position: plain.Position): Position;
/**
 * Converts from an instance of {@link Location} to the plain object {@link plain.Location}.
 *
 * @internal
 */
export declare function fromLocation(location: sourcegraph.Location): plain.Location;
/**
 * Converts from an instance of {@link Hover} to the plain object {@link plain.Hover}.
 *
 * @internal
 */
export declare function fromHover(hover: sourcegraph.Hover): plain.Hover;
/**
 * Converts from an instance of {@link Range} to the plain object {@link plain.Range}.
 *
 * @internal
 */
export declare function fromRange(range: Range | sourcegraph.Range | undefined): plain.Range | undefined;
