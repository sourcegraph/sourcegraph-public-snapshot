/**
 *
 * Tooltip thoughts:
 *
 *
 * Good implementation:
 *
 * <Button data-tooltip="Hello world">Test</Button>
 * BECOMES
 * <Tooltip title="Hello world">
 *  <Button>Test</Button>
 * </Tooltip>
 *
 *
 * Requirements:
 * 1. Show on short delay on hover (50ms?) = Needs implementing, delay on hover callback
 * 2. Show/hide immediately on focus/click = Needs implementing. Need to access inner element. Look at how Radix Trigger does it.
 * 3. Only ever show 1 tooltip at a time = Needs implementing. Global tooltip manager/context?
 * 4. Needs positioning/tail logic = Done through Popover.
 * 5. Needs to allow moving mouse towards tooltip. Easiest way seems to be to delay hiding the tooltip. = Needs implementing
 *
 */
