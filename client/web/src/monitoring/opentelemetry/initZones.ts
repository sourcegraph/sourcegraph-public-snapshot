// Use passive event listeners by default to improve perf and remove Chrome warning.
// https://github.com/angular/angular/blob/main/aio/content/guide/event-binding.md#binding-to-passive-events
// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
;(window as any).__zone_symbol__PASSIVE_EVENTS = ['scroll', 'touchstart', 'touchmove']
