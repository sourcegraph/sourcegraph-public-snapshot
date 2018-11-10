import { basename, dirname, extname } from 'path';
/**
 * Returns a new context created by applying the update context to the base context. It is equivalent to `{...base,
 * ...update}` in JavaScript except that null values in the update result in deletion of the property.
 */
export function applyContextUpdate(base, update) {
    const result = Object.assign({}, base);
    for (const [key, value] of Object.entries(update)) {
        if (value === null) {
            delete result[key];
        }
        else {
            result[key] = value;
        }
    }
    return result;
}
/** A context that has no properties. */
export const EMPTY_CONTEXT = {};
/**
 * Looks up a key in the computed context, which consists of special context properties (with higher precedence)
 * and the environment's context properties (with lower precedence).
 *
 * @param key the context property key to look up
 * @param scope the user interface component in whose scope this computation should occur
 */
export function getComputedContextProperty(environment, key, scope) {
    if (key.startsWith('config.')) {
        const prop = key.slice('config.'.length);
        const value = environment.configuration.merged[prop];
        // Map undefined to null because an undefined value is treated as "does not exist in
        // context" and an error is thrown, which is undesirable for config values (for
        // which a falsey null default is useful).
        return value === undefined ? null : value;
    }
    const textDocument = scope || (environment.visibleTextDocuments && environment.visibleTextDocuments[0]);
    if (key === 'resource' || key === 'component' /* BACKCOMPAT: allow 'component' */) {
        return !!textDocument;
    }
    if (key.startsWith('resource.')) {
        if (!textDocument) {
            return undefined;
        }
        // TODO(sqs): Define these precisely. If the resource is in a repository, what is the "path"? Is it the
        // path relative to the repository's root? If it's a file on disk, then "path" could also mean the
        // (absolute) path on the file system. Clear up that ambiguity.
        const prop = key.slice('resource.'.length);
        switch (prop) {
            case 'uri':
                return textDocument.uri;
            case 'basename':
                return basename(textDocument.uri);
            case 'dirname':
                return dirname(textDocument.uri);
            case 'extname':
                return extname(textDocument.uri);
            case 'language':
                return textDocument.languageId;
            case 'textContent':
                return textDocument.text;
            case 'type':
                return 'textDocument';
        }
    }
    if (key.startsWith('component.')) {
        if (!textDocument) {
            return undefined;
        }
        const prop = key.slice('component.'.length);
        switch (prop) {
            case 'type':
                return 'textEditor';
        }
    }
    if (key === 'context') {
        return environment.context;
    }
    return environment.context[key];
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY29udGV4dC5qcyIsInNvdXJjZVJvb3QiOiJzcmMvIiwic291cmNlcyI6WyJjbGllbnQvY29udGV4dC9jb250ZXh0LnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLE9BQU8sRUFBRSxRQUFRLEVBQUUsT0FBTyxFQUFFLE9BQU8sRUFBRSxNQUFNLE1BQU0sQ0FBQTtBQUlqRDs7O0dBR0c7QUFDSCxNQUFNLFVBQVUsa0JBQWtCLENBQUMsSUFBYSxFQUFFLE1BQWU7SUFDN0QsTUFBTSxNQUFNLHFCQUFRLElBQUksQ0FBRSxDQUFBO0lBQzFCLEtBQUssTUFBTSxDQUFDLEdBQUcsRUFBRSxLQUFLLENBQUMsSUFBSSxNQUFNLENBQUMsT0FBTyxDQUFDLE1BQU0sQ0FBQyxFQUFFO1FBQy9DLElBQUksS0FBSyxLQUFLLElBQUksRUFBRTtZQUNoQixPQUFPLE1BQU0sQ0FBQyxHQUFHLENBQUMsQ0FBQTtTQUNyQjthQUFNO1lBQ0gsTUFBTSxDQUFDLEdBQUcsQ0FBQyxHQUFHLEtBQUssQ0FBQTtTQUN0QjtLQUNKO0lBQ0QsT0FBTyxNQUFNLENBQUE7QUFDakIsQ0FBQztBQVNELHdDQUF3QztBQUN4QyxNQUFNLENBQUMsTUFBTSxhQUFhLEdBQVksRUFBRSxDQUFBO0FBRXhDOzs7Ozs7R0FNRztBQUNILE1BQU0sVUFBVSwwQkFBMEIsQ0FBQyxXQUF3QixFQUFFLEdBQVcsRUFBRSxLQUF3QjtJQUN0RyxJQUFJLEdBQUcsQ0FBQyxVQUFVLENBQUMsU0FBUyxDQUFDLEVBQUU7UUFDM0IsTUFBTSxJQUFJLEdBQUcsR0FBRyxDQUFDLEtBQUssQ0FBQyxTQUFTLENBQUMsTUFBTSxDQUFDLENBQUE7UUFDeEMsTUFBTSxLQUFLLEdBQUcsV0FBVyxDQUFDLGFBQWEsQ0FBQyxNQUFNLENBQUMsSUFBSSxDQUFDLENBQUE7UUFDcEQsb0ZBQW9GO1FBQ3BGLCtFQUErRTtRQUMvRSwwQ0FBMEM7UUFDMUMsT0FBTyxLQUFLLEtBQUssU0FBUyxDQUFDLENBQUMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLEtBQUssQ0FBQTtLQUM1QztJQUNELE1BQU0sWUFBWSxHQUNkLEtBQUssSUFBSSxDQUFDLFdBQVcsQ0FBQyxvQkFBb0IsSUFBSSxXQUFXLENBQUMsb0JBQW9CLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQTtJQUN0RixJQUFJLEdBQUcsS0FBSyxVQUFVLElBQUksR0FBRyxLQUFLLFdBQVcsQ0FBQyxtQ0FBbUMsRUFBRTtRQUMvRSxPQUFPLENBQUMsQ0FBQyxZQUFZLENBQUE7S0FDeEI7SUFDRCxJQUFJLEdBQUcsQ0FBQyxVQUFVLENBQUMsV0FBVyxDQUFDLEVBQUU7UUFDN0IsSUFBSSxDQUFDLFlBQVksRUFBRTtZQUNmLE9BQU8sU0FBUyxDQUFBO1NBQ25CO1FBQ0QsdUdBQXVHO1FBQ3ZHLGtHQUFrRztRQUNsRywrREFBK0Q7UUFDL0QsTUFBTSxJQUFJLEdBQUcsR0FBRyxDQUFDLEtBQUssQ0FBQyxXQUFXLENBQUMsTUFBTSxDQUFDLENBQUE7UUFDMUMsUUFBUSxJQUFJLEVBQUU7WUFDVixLQUFLLEtBQUs7Z0JBQ04sT0FBTyxZQUFZLENBQUMsR0FBRyxDQUFBO1lBQzNCLEtBQUssVUFBVTtnQkFDWCxPQUFPLFFBQVEsQ0FBQyxZQUFZLENBQUMsR0FBRyxDQUFDLENBQUE7WUFDckMsS0FBSyxTQUFTO2dCQUNWLE9BQU8sT0FBTyxDQUFDLFlBQVksQ0FBQyxHQUFHLENBQUMsQ0FBQTtZQUNwQyxLQUFLLFNBQVM7Z0JBQ1YsT0FBTyxPQUFPLENBQUMsWUFBWSxDQUFDLEdBQUcsQ0FBQyxDQUFBO1lBQ3BDLEtBQUssVUFBVTtnQkFDWCxPQUFPLFlBQVksQ0FBQyxVQUFVLENBQUE7WUFDbEMsS0FBSyxhQUFhO2dCQUNkLE9BQU8sWUFBWSxDQUFDLElBQUksQ0FBQTtZQUM1QixLQUFLLE1BQU07Z0JBQ1AsT0FBTyxjQUFjLENBQUE7U0FDNUI7S0FDSjtJQUNELElBQUksR0FBRyxDQUFDLFVBQVUsQ0FBQyxZQUFZLENBQUMsRUFBRTtRQUM5QixJQUFJLENBQUMsWUFBWSxFQUFFO1lBQ2YsT0FBTyxTQUFTLENBQUE7U0FDbkI7UUFDRCxNQUFNLElBQUksR0FBRyxHQUFHLENBQUMsS0FBSyxDQUFDLFlBQVksQ0FBQyxNQUFNLENBQUMsQ0FBQTtRQUMzQyxRQUFRLElBQUksRUFBRTtZQUNWLEtBQUssTUFBTTtnQkFDUCxPQUFPLFlBQVksQ0FBQTtTQUMxQjtLQUNKO0lBQ0QsSUFBSSxHQUFHLEtBQUssU0FBUyxFQUFFO1FBQ25CLE9BQU8sV0FBVyxDQUFDLE9BQU8sQ0FBQTtLQUM3QjtJQUNELE9BQU8sV0FBVyxDQUFDLE9BQU8sQ0FBQyxHQUFHLENBQUMsQ0FBQTtBQUNuQyxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgYmFzZW5hbWUsIGRpcm5hbWUsIGV4dG5hbWUgfSBmcm9tICdwYXRoJ1xuaW1wb3J0IHsgRW52aXJvbm1lbnQgfSBmcm9tICcuLi9lbnZpcm9ubWVudCdcbmltcG9ydCB7IFRleHREb2N1bWVudEl0ZW0gfSBmcm9tICcuLi90eXBlcy90ZXh0RG9jdW1lbnQnXG5cbi8qKlxuICogUmV0dXJucyBhIG5ldyBjb250ZXh0IGNyZWF0ZWQgYnkgYXBwbHlpbmcgdGhlIHVwZGF0ZSBjb250ZXh0IHRvIHRoZSBiYXNlIGNvbnRleHQuIEl0IGlzIGVxdWl2YWxlbnQgdG8gYHsuLi5iYXNlLFxuICogLi4udXBkYXRlfWAgaW4gSmF2YVNjcmlwdCBleGNlcHQgdGhhdCBudWxsIHZhbHVlcyBpbiB0aGUgdXBkYXRlIHJlc3VsdCBpbiBkZWxldGlvbiBvZiB0aGUgcHJvcGVydHkuXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBhcHBseUNvbnRleHRVcGRhdGUoYmFzZTogQ29udGV4dCwgdXBkYXRlOiBDb250ZXh0KTogQ29udGV4dCB7XG4gICAgY29uc3QgcmVzdWx0ID0geyAuLi5iYXNlIH1cbiAgICBmb3IgKGNvbnN0IFtrZXksIHZhbHVlXSBvZiBPYmplY3QuZW50cmllcyh1cGRhdGUpKSB7XG4gICAgICAgIGlmICh2YWx1ZSA9PT0gbnVsbCkge1xuICAgICAgICAgICAgZGVsZXRlIHJlc3VsdFtrZXldXG4gICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICByZXN1bHRba2V5XSA9IHZhbHVlXG4gICAgICAgIH1cbiAgICB9XG4gICAgcmV0dXJuIHJlc3VsdFxufVxuXG4vKipcbiAqIENvbnRleHQgaXMgYW4gYXJiaXRyYXJ5LCBpbW11dGFibGUgc2V0IG9mIGtleS12YWx1ZSBwYWlycy5cbiAqL1xuZXhwb3J0IGludGVyZmFjZSBDb250ZXh0IHtcbiAgICBba2V5OiBzdHJpbmddOiBzdHJpbmcgfCBudW1iZXIgfCBib29sZWFuIHwgQ29udGV4dCB8IG51bGxcbn1cblxuLyoqIEEgY29udGV4dCB0aGF0IGhhcyBubyBwcm9wZXJ0aWVzLiAqL1xuZXhwb3J0IGNvbnN0IEVNUFRZX0NPTlRFWFQ6IENvbnRleHQgPSB7fVxuXG4vKipcbiAqIExvb2tzIHVwIGEga2V5IGluIHRoZSBjb21wdXRlZCBjb250ZXh0LCB3aGljaCBjb25zaXN0cyBvZiBzcGVjaWFsIGNvbnRleHQgcHJvcGVydGllcyAod2l0aCBoaWdoZXIgcHJlY2VkZW5jZSlcbiAqIGFuZCB0aGUgZW52aXJvbm1lbnQncyBjb250ZXh0IHByb3BlcnRpZXMgKHdpdGggbG93ZXIgcHJlY2VkZW5jZSkuXG4gKlxuICogQHBhcmFtIGtleSB0aGUgY29udGV4dCBwcm9wZXJ0eSBrZXkgdG8gbG9vayB1cFxuICogQHBhcmFtIHNjb3BlIHRoZSB1c2VyIGludGVyZmFjZSBjb21wb25lbnQgaW4gd2hvc2Ugc2NvcGUgdGhpcyBjb21wdXRhdGlvbiBzaG91bGQgb2NjdXJcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIGdldENvbXB1dGVkQ29udGV4dFByb3BlcnR5KGVudmlyb25tZW50OiBFbnZpcm9ubWVudCwga2V5OiBzdHJpbmcsIHNjb3BlPzogVGV4dERvY3VtZW50SXRlbSk6IGFueSB7XG4gICAgaWYgKGtleS5zdGFydHNXaXRoKCdjb25maWcuJykpIHtcbiAgICAgICAgY29uc3QgcHJvcCA9IGtleS5zbGljZSgnY29uZmlnLicubGVuZ3RoKVxuICAgICAgICBjb25zdCB2YWx1ZSA9IGVudmlyb25tZW50LmNvbmZpZ3VyYXRpb24ubWVyZ2VkW3Byb3BdXG4gICAgICAgIC8vIE1hcCB1bmRlZmluZWQgdG8gbnVsbCBiZWNhdXNlIGFuIHVuZGVmaW5lZCB2YWx1ZSBpcyB0cmVhdGVkIGFzIFwiZG9lcyBub3QgZXhpc3QgaW5cbiAgICAgICAgLy8gY29udGV4dFwiIGFuZCBhbiBlcnJvciBpcyB0aHJvd24sIHdoaWNoIGlzIHVuZGVzaXJhYmxlIGZvciBjb25maWcgdmFsdWVzIChmb3JcbiAgICAgICAgLy8gd2hpY2ggYSBmYWxzZXkgbnVsbCBkZWZhdWx0IGlzIHVzZWZ1bCkuXG4gICAgICAgIHJldHVybiB2YWx1ZSA9PT0gdW5kZWZpbmVkID8gbnVsbCA6IHZhbHVlXG4gICAgfVxuICAgIGNvbnN0IHRleHREb2N1bWVudDogVGV4dERvY3VtZW50SXRlbSB8IG51bGwgPVxuICAgICAgICBzY29wZSB8fCAoZW52aXJvbm1lbnQudmlzaWJsZVRleHREb2N1bWVudHMgJiYgZW52aXJvbm1lbnQudmlzaWJsZVRleHREb2N1bWVudHNbMF0pXG4gICAgaWYgKGtleSA9PT0gJ3Jlc291cmNlJyB8fCBrZXkgPT09ICdjb21wb25lbnQnIC8qIEJBQ0tDT01QQVQ6IGFsbG93ICdjb21wb25lbnQnICovKSB7XG4gICAgICAgIHJldHVybiAhIXRleHREb2N1bWVudFxuICAgIH1cbiAgICBpZiAoa2V5LnN0YXJ0c1dpdGgoJ3Jlc291cmNlLicpKSB7XG4gICAgICAgIGlmICghdGV4dERvY3VtZW50KSB7XG4gICAgICAgICAgICByZXR1cm4gdW5kZWZpbmVkXG4gICAgICAgIH1cbiAgICAgICAgLy8gVE9ETyhzcXMpOiBEZWZpbmUgdGhlc2UgcHJlY2lzZWx5LiBJZiB0aGUgcmVzb3VyY2UgaXMgaW4gYSByZXBvc2l0b3J5LCB3aGF0IGlzIHRoZSBcInBhdGhcIj8gSXMgaXQgdGhlXG4gICAgICAgIC8vIHBhdGggcmVsYXRpdmUgdG8gdGhlIHJlcG9zaXRvcnkncyByb290PyBJZiBpdCdzIGEgZmlsZSBvbiBkaXNrLCB0aGVuIFwicGF0aFwiIGNvdWxkIGFsc28gbWVhbiB0aGVcbiAgICAgICAgLy8gKGFic29sdXRlKSBwYXRoIG9uIHRoZSBmaWxlIHN5c3RlbS4gQ2xlYXIgdXAgdGhhdCBhbWJpZ3VpdHkuXG4gICAgICAgIGNvbnN0IHByb3AgPSBrZXkuc2xpY2UoJ3Jlc291cmNlLicubGVuZ3RoKVxuICAgICAgICBzd2l0Y2ggKHByb3ApIHtcbiAgICAgICAgICAgIGNhc2UgJ3VyaSc6XG4gICAgICAgICAgICAgICAgcmV0dXJuIHRleHREb2N1bWVudC51cmlcbiAgICAgICAgICAgIGNhc2UgJ2Jhc2VuYW1lJzpcbiAgICAgICAgICAgICAgICByZXR1cm4gYmFzZW5hbWUodGV4dERvY3VtZW50LnVyaSlcbiAgICAgICAgICAgIGNhc2UgJ2Rpcm5hbWUnOlxuICAgICAgICAgICAgICAgIHJldHVybiBkaXJuYW1lKHRleHREb2N1bWVudC51cmkpXG4gICAgICAgICAgICBjYXNlICdleHRuYW1lJzpcbiAgICAgICAgICAgICAgICByZXR1cm4gZXh0bmFtZSh0ZXh0RG9jdW1lbnQudXJpKVxuICAgICAgICAgICAgY2FzZSAnbGFuZ3VhZ2UnOlxuICAgICAgICAgICAgICAgIHJldHVybiB0ZXh0RG9jdW1lbnQubGFuZ3VhZ2VJZFxuICAgICAgICAgICAgY2FzZSAndGV4dENvbnRlbnQnOlxuICAgICAgICAgICAgICAgIHJldHVybiB0ZXh0RG9jdW1lbnQudGV4dFxuICAgICAgICAgICAgY2FzZSAndHlwZSc6XG4gICAgICAgICAgICAgICAgcmV0dXJuICd0ZXh0RG9jdW1lbnQnXG4gICAgICAgIH1cbiAgICB9XG4gICAgaWYgKGtleS5zdGFydHNXaXRoKCdjb21wb25lbnQuJykpIHtcbiAgICAgICAgaWYgKCF0ZXh0RG9jdW1lbnQpIHtcbiAgICAgICAgICAgIHJldHVybiB1bmRlZmluZWRcbiAgICAgICAgfVxuICAgICAgICBjb25zdCBwcm9wID0ga2V5LnNsaWNlKCdjb21wb25lbnQuJy5sZW5ndGgpXG4gICAgICAgIHN3aXRjaCAocHJvcCkge1xuICAgICAgICAgICAgY2FzZSAndHlwZSc6XG4gICAgICAgICAgICAgICAgcmV0dXJuICd0ZXh0RWRpdG9yJ1xuICAgICAgICB9XG4gICAgfVxuICAgIGlmIChrZXkgPT09ICdjb250ZXh0Jykge1xuICAgICAgICByZXR1cm4gZW52aXJvbm1lbnQuY29udGV4dFxuICAgIH1cbiAgICByZXR1cm4gZW52aXJvbm1lbnQuY29udGV4dFtrZXldXG59XG4iXX0=