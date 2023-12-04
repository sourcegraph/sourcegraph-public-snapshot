import { announce, clearAnnouncer } from '@react-aria/live-announcer'

/**
 * Announces a message to screen readers.
 *
 * This should be used for any information that is not visible on the screen and cannot be inferred by a screen reader.
 * For example: A submitted modal automatically closing is not an obvious "success" to a screen reader user.
 * We should use this utility to ensure those users are still informed that something happened.
 */
export const screenReaderAnnounce = announce

/**
 * Clears any queued screen reader announcements.
 *
 * This should be only be used if a user does something that should interrupt a previous announcement (like cancelling an action).
 * Use sparingly, it can cause a confusing UX and generally shouldn't be required.
 */
export const screenReaderClearAnnouncements = clearAnnouncer
