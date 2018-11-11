export var HoverMerged;
(function (HoverMerged) {
    /** Create a merged hover from the given individual hovers. */
    function from(values) {
        const contents = [];
        let range;
        for (const result of values) {
            if (result) {
                if (result.contents && result.contents.value) {
                    contents.push({
                        value: result.contents.value,
                        kind: result.contents.kind || 'plaintext',
                    });
                }
                const __backcompatContents = result.__backcompatContents; // tslint:disable-line deprecation
                if (__backcompatContents) {
                    for (const content of Array.isArray(__backcompatContents)
                        ? __backcompatContents
                        : [__backcompatContents]) {
                        if (typeof content === 'string') {
                            if (content) {
                                contents.push(content);
                            }
                        }
                        else if ('language' in content) {
                            if (content.language && content.value) {
                                contents.push(content);
                            }
                        }
                        else if ('value' in content) {
                            if (content.value) {
                                contents.push(content.value);
                            }
                        }
                    }
                }
                if (result.range && !range) {
                    range = result.range;
                }
            }
        }
        return contents.length === 0 ? null : range ? { contents, range } : { contents };
    }
    HoverMerged.from = from;
})(HoverMerged || (HoverMerged = {}));
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaG92ZXIuanMiLCJzb3VyY2VSb290Ijoic3JjLyIsInNvdXJjZXMiOlsiY2xpZW50L3R5cGVzL2hvdmVyLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQWlCQSxNQUFNLEtBQVcsV0FBVyxDQXdDM0I7QUF4Q0QsV0FBaUIsV0FBVztJQUN4Qiw4REFBOEQ7SUFDOUQsU0FBZ0IsSUFBSSxDQUFDLE1BQWlEO1FBQ2xFLE1BQU0sUUFBUSxHQUE0QixFQUFFLENBQUE7UUFDNUMsSUFBSSxLQUEyQixDQUFBO1FBQy9CLEtBQUssTUFBTSxNQUFNLElBQUksTUFBTSxFQUFFO1lBQ3pCLElBQUksTUFBTSxFQUFFO2dCQUNSLElBQUksTUFBTSxDQUFDLFFBQVEsSUFBSSxNQUFNLENBQUMsUUFBUSxDQUFDLEtBQUssRUFBRTtvQkFDMUMsUUFBUSxDQUFDLElBQUksQ0FBQzt3QkFDVixLQUFLLEVBQUUsTUFBTSxDQUFDLFFBQVEsQ0FBQyxLQUFLO3dCQUM1QixJQUFJLEVBQUUsTUFBTSxDQUFDLFFBQVEsQ0FBQyxJQUFJLElBQUssV0FBMEI7cUJBQzVELENBQUMsQ0FBQTtpQkFDTDtnQkFDRCxNQUFNLG9CQUFvQixHQUFHLE1BQU0sQ0FBQyxvQkFBb0IsQ0FBQSxDQUFDLGtDQUFrQztnQkFDM0YsSUFBSSxvQkFBb0IsRUFBRTtvQkFDdEIsS0FBSyxNQUFNLE9BQU8sSUFBSSxLQUFLLENBQUMsT0FBTyxDQUFDLG9CQUFvQixDQUFDO3dCQUNyRCxDQUFDLENBQUMsb0JBQW9CO3dCQUN0QixDQUFDLENBQUMsQ0FBQyxvQkFBb0IsQ0FBQyxFQUFFO3dCQUMxQixJQUFJLE9BQU8sT0FBTyxLQUFLLFFBQVEsRUFBRTs0QkFDN0IsSUFBSSxPQUFPLEVBQUU7Z0NBQ1QsUUFBUSxDQUFDLElBQUksQ0FBQyxPQUFPLENBQUMsQ0FBQTs2QkFDekI7eUJBQ0o7NkJBQU0sSUFBSSxVQUFVLElBQUksT0FBTyxFQUFFOzRCQUM5QixJQUFJLE9BQU8sQ0FBQyxRQUFRLElBQUksT0FBTyxDQUFDLEtBQUssRUFBRTtnQ0FDbkMsUUFBUSxDQUFDLElBQUksQ0FBQyxPQUFPLENBQUMsQ0FBQTs2QkFDekI7eUJBQ0o7NkJBQU0sSUFBSSxPQUFPLElBQUksT0FBTyxFQUFFOzRCQUMzQixJQUFJLE9BQU8sQ0FBQyxLQUFLLEVBQUU7Z0NBQ2YsUUFBUSxDQUFDLElBQUksQ0FBQyxPQUFPLENBQUMsS0FBSyxDQUFDLENBQUE7NkJBQy9CO3lCQUNKO3FCQUNKO2lCQUNKO2dCQUNELElBQUksTUFBTSxDQUFDLEtBQUssSUFBSSxDQUFDLEtBQUssRUFBRTtvQkFDeEIsS0FBSyxHQUFHLE1BQU0sQ0FBQyxLQUFLLENBQUE7aUJBQ3ZCO2FBQ0o7U0FDSjtRQUNELE9BQU8sUUFBUSxDQUFDLE1BQU0sS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsS0FBSyxDQUFDLENBQUMsQ0FBQyxFQUFFLFFBQVEsRUFBRSxLQUFLLEVBQUUsQ0FBQyxDQUFDLENBQUMsRUFBRSxRQUFRLEVBQUUsQ0FBQTtJQUNwRixDQUFDO0lBckNlLGdCQUFJLE9BcUNuQixDQUFBO0FBQ0wsQ0FBQyxFQXhDZ0IsV0FBVyxLQUFYLFdBQVcsUUF3QzNCIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgSG92ZXIsIE1hcmt1cENvbnRlbnQsIE1hcmt1cEtpbmQgfSBmcm9tICdzb3VyY2VncmFwaCdcbmltcG9ydCB7IEhvdmVyIGFzIFBsYWluSG92ZXIsIFJhbmdlIH0gZnJvbSAnLi4vLi4vcHJvdG9jb2wvcGxhaW5UeXBlcydcblxuLyoqIEEgaG92ZXIgdGhhdCBpcyBtZXJnZWQgZnJvbSBtdWx0aXBsZSBIb3ZlciByZXN1bHRzIGFuZCBub3JtYWxpemVkLiAqL1xuZXhwb3J0IGludGVyZmFjZSBIb3Zlck1lcmdlZCB7XG4gICAgLyoqXG4gICAgICogQHRvZG8gTWFrZSB0aGlzIHR5cGUgKmp1c3QqIHtAbGluayBNYXJrdXBDb250ZW50fSB3aGVuIGFsbCBjb25zdW1lcnMgYXJlIHVwZGF0ZWQuXG4gICAgICovXG4gICAgY29udGVudHM6XG4gICAgICAgIHwgTWFya3VwQ29udGVudFxuICAgICAgICB8IHN0cmluZ1xuICAgICAgICB8IHsgbGFuZ3VhZ2U6IHN0cmluZzsgdmFsdWU6IHN0cmluZyB9XG4gICAgICAgIHwgKE1hcmt1cENvbnRlbnQgfCBzdHJpbmcgfCB7IGxhbmd1YWdlOiBzdHJpbmc7IHZhbHVlOiBzdHJpbmcgfSlbXVxuXG4gICAgcmFuZ2U/OiBSYW5nZVxufVxuXG5leHBvcnQgbmFtZXNwYWNlIEhvdmVyTWVyZ2VkIHtcbiAgICAvKiogQ3JlYXRlIGEgbWVyZ2VkIGhvdmVyIGZyb20gdGhlIGdpdmVuIGluZGl2aWR1YWwgaG92ZXJzLiAqL1xuICAgIGV4cG9ydCBmdW5jdGlvbiBmcm9tKHZhbHVlczogKEhvdmVyIHwgUGxhaW5Ib3ZlciB8IG51bGwgfCB1bmRlZmluZWQpW10pOiBIb3Zlck1lcmdlZCB8IG51bGwge1xuICAgICAgICBjb25zdCBjb250ZW50czogSG92ZXJNZXJnZWRbJ2NvbnRlbnRzJ10gPSBbXVxuICAgICAgICBsZXQgcmFuZ2U6IEhvdmVyTWVyZ2VkWydyYW5nZSddXG4gICAgICAgIGZvciAoY29uc3QgcmVzdWx0IG9mIHZhbHVlcykge1xuICAgICAgICAgICAgaWYgKHJlc3VsdCkge1xuICAgICAgICAgICAgICAgIGlmIChyZXN1bHQuY29udGVudHMgJiYgcmVzdWx0LmNvbnRlbnRzLnZhbHVlKSB7XG4gICAgICAgICAgICAgICAgICAgIGNvbnRlbnRzLnB1c2goe1xuICAgICAgICAgICAgICAgICAgICAgICAgdmFsdWU6IHJlc3VsdC5jb250ZW50cy52YWx1ZSxcbiAgICAgICAgICAgICAgICAgICAgICAgIGtpbmQ6IHJlc3VsdC5jb250ZW50cy5raW5kIHx8ICgncGxhaW50ZXh0JyBhcyBNYXJrdXBLaW5kKSxcbiAgICAgICAgICAgICAgICAgICAgfSlcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgY29uc3QgX19iYWNrY29tcGF0Q29udGVudHMgPSByZXN1bHQuX19iYWNrY29tcGF0Q29udGVudHMgLy8gdHNsaW50OmRpc2FibGUtbGluZSBkZXByZWNhdGlvblxuICAgICAgICAgICAgICAgIGlmIChfX2JhY2tjb21wYXRDb250ZW50cykge1xuICAgICAgICAgICAgICAgICAgICBmb3IgKGNvbnN0IGNvbnRlbnQgb2YgQXJyYXkuaXNBcnJheShfX2JhY2tjb21wYXRDb250ZW50cylcbiAgICAgICAgICAgICAgICAgICAgICAgID8gX19iYWNrY29tcGF0Q29udGVudHNcbiAgICAgICAgICAgICAgICAgICAgICAgIDogW19fYmFja2NvbXBhdENvbnRlbnRzXSkge1xuICAgICAgICAgICAgICAgICAgICAgICAgaWYgKHR5cGVvZiBjb250ZW50ID09PSAnc3RyaW5nJykge1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgIGlmIChjb250ZW50KSB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIGNvbnRlbnRzLnB1c2goY29udGVudClcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgICAgICAgICB9IGVsc2UgaWYgKCdsYW5ndWFnZScgaW4gY29udGVudCkge1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgIGlmIChjb250ZW50Lmxhbmd1YWdlICYmIGNvbnRlbnQudmFsdWUpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgY29udGVudHMucHVzaChjb250ZW50KVxuICAgICAgICAgICAgICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICAgICAgICAgIH0gZWxzZSBpZiAoJ3ZhbHVlJyBpbiBjb250ZW50KSB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgaWYgKGNvbnRlbnQudmFsdWUpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgY29udGVudHMucHVzaChjb250ZW50LnZhbHVlKVxuICAgICAgICAgICAgICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICBpZiAocmVzdWx0LnJhbmdlICYmICFyYW5nZSkge1xuICAgICAgICAgICAgICAgICAgICByYW5nZSA9IHJlc3VsdC5yYW5nZVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgfVxuICAgICAgICByZXR1cm4gY29udGVudHMubGVuZ3RoID09PSAwID8gbnVsbCA6IHJhbmdlID8geyBjb250ZW50cywgcmFuZ2UgfSA6IHsgY29udGVudHMgfVxuICAgIH1cbn1cbiJdfQ==