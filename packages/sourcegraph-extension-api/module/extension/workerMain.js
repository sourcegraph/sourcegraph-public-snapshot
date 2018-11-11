import { tryCatchPromise } from '../util';
import { createExtensionHost } from './extensionHost';
/**
 * The entrypoint for Web Workers that are spawned to run an extension.
 *
 * To initialize the worker, the parent sends it a message whose data is an object conforming to the
 * {@link InitData} interface. Among other things, this contains the URL of the extension's JavaScript bundle.
 *
 * @param self The worker's `self` global scope.
 */
export function extensionHostWorkerMain(self) {
    self.addEventListener('message', receiveExtensionURL);
    function receiveExtensionURL(ev) {
        try {
            // Only listen for the 1st URL.
            self.removeEventListener('message', receiveExtensionURL);
            if (ev.origin && ev.origin !== self.location.origin) {
                console.error(`Invalid extension host message origin: ${ev.origin} (expected ${self.location.origin})`);
                self.close();
            }
            const initData = ev.data;
            if (typeof initData.bundleURL !== 'string' || !initData.bundleURL.startsWith('blob:')) {
                console.error(`Invalid extension bundle URL: ${initData.bundleURL}`);
                self.close();
            }
            const api = createExtensionHost(initData);
            self.require = (modulePath) => {
                if (modulePath === 'sourcegraph') {
                    return api;
                }
                throw new Error(`require: module not found: ${modulePath}`);
            };
            self.exports = {};
            self.module = {};
            self.importScripts(initData.bundleURL);
            const extensionExports = self.module.exports;
            delete self.module;
            if ('activate' in extensionExports) {
                try {
                    tryCatchPromise(() => extensionExports.activate()).catch((err) => {
                        console.error(`Error creating extension host:`, err);
                        self.close();
                    });
                }
                catch (err) {
                    console.error(`Error activating extension.`, err);
                }
            }
            else {
                console.error(`Extension did not export an 'activate' function.`);
            }
        }
        catch (err) {
            console.error(err);
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoid29ya2VyTWFpbi5qcyIsInNvdXJjZVJvb3QiOiJzcmMvIiwic291cmNlcyI6WyJleHRlbnNpb24vd29ya2VyTWFpbi50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxPQUFPLEVBQUUsZUFBZSxFQUFFLE1BQU0sU0FBUyxDQUFBO0FBQ3pDLE9BQU8sRUFBRSxtQkFBbUIsRUFBWSxNQUFNLGlCQUFpQixDQUFBO0FBcUIvRDs7Ozs7OztHQU9HO0FBQ0gsTUFBTSxVQUFVLHVCQUF1QixDQUFDLElBQWdDO0lBQ3BFLElBQUksQ0FBQyxnQkFBZ0IsQ0FBQyxTQUFTLEVBQUUsbUJBQW1CLENBQUMsQ0FBQTtJQUVyRCxTQUFTLG1CQUFtQixDQUFDLEVBQWdCO1FBQ3pDLElBQUk7WUFDQSwrQkFBK0I7WUFDL0IsSUFBSSxDQUFDLG1CQUFtQixDQUFDLFNBQVMsRUFBRSxtQkFBbUIsQ0FBQyxDQUFBO1lBRXhELElBQUksRUFBRSxDQUFDLE1BQU0sSUFBSSxFQUFFLENBQUMsTUFBTSxLQUFLLElBQUksQ0FBQyxRQUFRLENBQUMsTUFBTSxFQUFFO2dCQUNqRCxPQUFPLENBQUMsS0FBSyxDQUFDLDBDQUEwQyxFQUFFLENBQUMsTUFBTSxjQUFjLElBQUksQ0FBQyxRQUFRLENBQUMsTUFBTSxHQUFHLENBQUMsQ0FBQTtnQkFDdkcsSUFBSSxDQUFDLEtBQUssRUFBRSxDQUFBO2FBQ2Y7WUFFRCxNQUFNLFFBQVEsR0FBYSxFQUFFLENBQUMsSUFBSSxDQUFBO1lBQ2xDLElBQUksT0FBTyxRQUFRLENBQUMsU0FBUyxLQUFLLFFBQVEsSUFBSSxDQUFDLFFBQVEsQ0FBQyxTQUFTLENBQUMsVUFBVSxDQUFDLE9BQU8sQ0FBQyxFQUFFO2dCQUNuRixPQUFPLENBQUMsS0FBSyxDQUFDLGlDQUFpQyxRQUFRLENBQUMsU0FBUyxFQUFFLENBQUMsQ0FBQTtnQkFDcEUsSUFBSSxDQUFDLEtBQUssRUFBRSxDQUFBO2FBQ2Y7WUFFRCxNQUFNLEdBQUcsR0FBRyxtQkFBbUIsQ0FBQyxRQUFRLENBQUMsQ0FHeEM7WUFBQyxJQUFZLENBQUMsT0FBTyxHQUFHLENBQUMsVUFBa0IsRUFBTyxFQUFFO2dCQUNqRCxJQUFJLFVBQVUsS0FBSyxhQUFhLEVBQUU7b0JBQzlCLE9BQU8sR0FBRyxDQUFBO2lCQUNiO2dCQUNELE1BQU0sSUFBSSxLQUFLLENBQUMsOEJBQThCLFVBQVUsRUFBRSxDQUFDLENBQUE7WUFDL0QsQ0FBQyxDQUlBO1lBQUMsSUFBWSxDQUFDLE9BQU8sR0FBRyxFQUFFLENBQzFCO1lBQUMsSUFBWSxDQUFDLE1BQU0sR0FBRyxFQUFFLENBQUE7WUFDMUIsSUFBSSxDQUFDLGFBQWEsQ0FBQyxRQUFRLENBQUMsU0FBUyxDQUFDLENBQUE7WUFDdEMsTUFBTSxnQkFBZ0IsR0FBSSxJQUFZLENBQUMsTUFBTSxDQUFDLE9BQU8sQ0FBQTtZQUNyRCxPQUFRLElBQVksQ0FBQyxNQUFNLENBQUE7WUFFM0IsSUFBSSxVQUFVLElBQUksZ0JBQWdCLEVBQUU7Z0JBQ2hDLElBQUk7b0JBQ0EsZUFBZSxDQUFDLEdBQUcsRUFBRSxDQUFDLGdCQUFnQixDQUFDLFFBQVEsRUFBRSxDQUFDLENBQUMsS0FBSyxDQUFDLENBQUMsR0FBUSxFQUFFLEVBQUU7d0JBQ2xFLE9BQU8sQ0FBQyxLQUFLLENBQUMsZ0NBQWdDLEVBQUUsR0FBRyxDQUFDLENBQUE7d0JBQ3BELElBQUksQ0FBQyxLQUFLLEVBQUUsQ0FBQTtvQkFDaEIsQ0FBQyxDQUFDLENBQUE7aUJBQ0w7Z0JBQUMsT0FBTyxHQUFHLEVBQUU7b0JBQ1YsT0FBTyxDQUFDLEtBQUssQ0FBQyw2QkFBNkIsRUFBRSxHQUFHLENBQUMsQ0FBQTtpQkFDcEQ7YUFDSjtpQkFBTTtnQkFDSCxPQUFPLENBQUMsS0FBSyxDQUFDLGtEQUFrRCxDQUFDLENBQUE7YUFDcEU7U0FDSjtRQUFDLE9BQU8sR0FBRyxFQUFFO1lBQ1YsT0FBTyxDQUFDLEtBQUssQ0FBQyxHQUFHLENBQUMsQ0FBQTtTQUNyQjtJQUNMLENBQUM7QUFDTCxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgdHJ5Q2F0Y2hQcm9taXNlIH0gZnJvbSAnLi4vdXRpbCdcbmltcG9ydCB7IGNyZWF0ZUV4dGVuc2lvbkhvc3QsIEluaXREYXRhIH0gZnJvbSAnLi9leHRlbnNpb25Ib3N0J1xuXG5pbnRlcmZhY2UgTWVzc2FnZUV2ZW50IHtcbiAgICBkYXRhOiBhbnlcbiAgICBvcmlnaW46IHN0cmluZyB8IG51bGxcbn1cblxuLyoqXG4gKiBUaGlzIGlzIGEgc3Vic2V0IG9mIERlZGljYXRlZFdvcmtlckdsb2JhbFNjb3BlLiBXZSBjYW4ndCB1c2UgYC8vLyA8cmVmZXJlbmNlcyBsaWI9XCJ3ZWJ3b3JrZXJcIi8+YCBiZWNhdXNlXG4gKiBQcmV0dGllciBkb2VzIG5vdCBzdXBwb3J0IHRyaXBsZS1zbGFzaCBkaXJlY3RpdmUgc3ludGF4LlxuICovXG5pbnRlcmZhY2UgRGVkaWNhdGVkV29ya2VyR2xvYmFsU2NvcGUge1xuICAgIGxvY2F0aW9uOiB7XG4gICAgICAgIG9yaWdpbjogc3RyaW5nXG4gICAgfVxuICAgIGFkZEV2ZW50TGlzdGVuZXIodHlwZTogJ21lc3NhZ2UnLCBsaXN0ZW5lcjogKGV2ZW50OiBNZXNzYWdlRXZlbnQpID0+IHZvaWQpOiB2b2lkXG4gICAgcmVtb3ZlRXZlbnRMaXN0ZW5lcih0eXBlOiAnbWVzc2FnZScsIGxpc3RlbmVyOiAoZXZlbnQ6IE1lc3NhZ2VFdmVudCkgPT4gdm9pZCk6IHZvaWRcbiAgICBpbXBvcnRTY3JpcHRzKHVybDogc3RyaW5nKTogdm9pZFxuICAgIGNsb3NlKCk6IHZvaWRcbn1cblxuLyoqXG4gKiBUaGUgZW50cnlwb2ludCBmb3IgV2ViIFdvcmtlcnMgdGhhdCBhcmUgc3Bhd25lZCB0byBydW4gYW4gZXh0ZW5zaW9uLlxuICpcbiAqIFRvIGluaXRpYWxpemUgdGhlIHdvcmtlciwgdGhlIHBhcmVudCBzZW5kcyBpdCBhIG1lc3NhZ2Ugd2hvc2UgZGF0YSBpcyBhbiBvYmplY3QgY29uZm9ybWluZyB0byB0aGVcbiAqIHtAbGluayBJbml0RGF0YX0gaW50ZXJmYWNlLiBBbW9uZyBvdGhlciB0aGluZ3MsIHRoaXMgY29udGFpbnMgdGhlIFVSTCBvZiB0aGUgZXh0ZW5zaW9uJ3MgSmF2YVNjcmlwdCBidW5kbGUuXG4gKlxuICogQHBhcmFtIHNlbGYgVGhlIHdvcmtlcidzIGBzZWxmYCBnbG9iYWwgc2NvcGUuXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBleHRlbnNpb25Ib3N0V29ya2VyTWFpbihzZWxmOiBEZWRpY2F0ZWRXb3JrZXJHbG9iYWxTY29wZSk6IHZvaWQge1xuICAgIHNlbGYuYWRkRXZlbnRMaXN0ZW5lcignbWVzc2FnZScsIHJlY2VpdmVFeHRlbnNpb25VUkwpXG5cbiAgICBmdW5jdGlvbiByZWNlaXZlRXh0ZW5zaW9uVVJMKGV2OiBNZXNzYWdlRXZlbnQpOiB2b2lkIHtcbiAgICAgICAgdHJ5IHtcbiAgICAgICAgICAgIC8vIE9ubHkgbGlzdGVuIGZvciB0aGUgMXN0IFVSTC5cbiAgICAgICAgICAgIHNlbGYucmVtb3ZlRXZlbnRMaXN0ZW5lcignbWVzc2FnZScsIHJlY2VpdmVFeHRlbnNpb25VUkwpXG5cbiAgICAgICAgICAgIGlmIChldi5vcmlnaW4gJiYgZXYub3JpZ2luICE9PSBzZWxmLmxvY2F0aW9uLm9yaWdpbikge1xuICAgICAgICAgICAgICAgIGNvbnNvbGUuZXJyb3IoYEludmFsaWQgZXh0ZW5zaW9uIGhvc3QgbWVzc2FnZSBvcmlnaW46ICR7ZXYub3JpZ2lufSAoZXhwZWN0ZWQgJHtzZWxmLmxvY2F0aW9uLm9yaWdpbn0pYClcbiAgICAgICAgICAgICAgICBzZWxmLmNsb3NlKClcbiAgICAgICAgICAgIH1cblxuICAgICAgICAgICAgY29uc3QgaW5pdERhdGE6IEluaXREYXRhID0gZXYuZGF0YVxuICAgICAgICAgICAgaWYgKHR5cGVvZiBpbml0RGF0YS5idW5kbGVVUkwgIT09ICdzdHJpbmcnIHx8ICFpbml0RGF0YS5idW5kbGVVUkwuc3RhcnRzV2l0aCgnYmxvYjonKSkge1xuICAgICAgICAgICAgICAgIGNvbnNvbGUuZXJyb3IoYEludmFsaWQgZXh0ZW5zaW9uIGJ1bmRsZSBVUkw6ICR7aW5pdERhdGEuYnVuZGxlVVJMfWApXG4gICAgICAgICAgICAgICAgc2VsZi5jbG9zZSgpXG4gICAgICAgICAgICB9XG5cbiAgICAgICAgICAgIGNvbnN0IGFwaSA9IGNyZWF0ZUV4dGVuc2lvbkhvc3QoaW5pdERhdGEpXG4gICAgICAgICAgICAvLyBNYWtlIGBpbXBvcnQgJ3NvdXJjZWdyYXBoJ2Agb3IgYHJlcXVpcmUoJ3NvdXJjZWdyYXBoJylgIHJldHVybiB0aGUgZXh0ZW5zaW9uIGhvc3Qnc1xuICAgICAgICAgICAgLy8gaW1wbGVtZW50YXRpb24gb2YgdGhlIGBzb3VyY2VncmFwaGAgbW9kdWxlLlxuICAgICAgICAgICAgOyhzZWxmIGFzIGFueSkucmVxdWlyZSA9IChtb2R1bGVQYXRoOiBzdHJpbmcpOiBhbnkgPT4ge1xuICAgICAgICAgICAgICAgIGlmIChtb2R1bGVQYXRoID09PSAnc291cmNlZ3JhcGgnKSB7XG4gICAgICAgICAgICAgICAgICAgIHJldHVybiBhcGlcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgdGhyb3cgbmV3IEVycm9yKGByZXF1aXJlOiBtb2R1bGUgbm90IGZvdW5kOiAke21vZHVsZVBhdGh9YClcbiAgICAgICAgICAgIH1cblxuICAgICAgICAgICAgLy8gTG9hZCB0aGUgZXh0ZW5zaW9uIGJ1bmRsZSBhbmQgcmV0cmlldmUgdGhlIGV4dGVuc2lvbiBlbnRyeXBvaW50IG1vZHVsZSdzIGV4cG9ydHMgb24gdGhlIGdsb2JhbFxuICAgICAgICAgICAgLy8gYG1vZHVsZWAgcHJvcGVydHkuXG4gICAgICAgICAgICA7KHNlbGYgYXMgYW55KS5leHBvcnRzID0ge31cbiAgICAgICAgICAgIDsoc2VsZiBhcyBhbnkpLm1vZHVsZSA9IHt9XG4gICAgICAgICAgICBzZWxmLmltcG9ydFNjcmlwdHMoaW5pdERhdGEuYnVuZGxlVVJMKVxuICAgICAgICAgICAgY29uc3QgZXh0ZW5zaW9uRXhwb3J0cyA9IChzZWxmIGFzIGFueSkubW9kdWxlLmV4cG9ydHNcbiAgICAgICAgICAgIGRlbGV0ZSAoc2VsZiBhcyBhbnkpLm1vZHVsZVxuXG4gICAgICAgICAgICBpZiAoJ2FjdGl2YXRlJyBpbiBleHRlbnNpb25FeHBvcnRzKSB7XG4gICAgICAgICAgICAgICAgdHJ5IHtcbiAgICAgICAgICAgICAgICAgICAgdHJ5Q2F0Y2hQcm9taXNlKCgpID0+IGV4dGVuc2lvbkV4cG9ydHMuYWN0aXZhdGUoKSkuY2F0Y2goKGVycjogYW55KSA9PiB7XG4gICAgICAgICAgICAgICAgICAgICAgICBjb25zb2xlLmVycm9yKGBFcnJvciBjcmVhdGluZyBleHRlbnNpb24gaG9zdDpgLCBlcnIpXG4gICAgICAgICAgICAgICAgICAgICAgICBzZWxmLmNsb3NlKClcbiAgICAgICAgICAgICAgICAgICAgfSlcbiAgICAgICAgICAgICAgICB9IGNhdGNoIChlcnIpIHtcbiAgICAgICAgICAgICAgICAgICAgY29uc29sZS5lcnJvcihgRXJyb3IgYWN0aXZhdGluZyBleHRlbnNpb24uYCwgZXJyKVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgY29uc29sZS5lcnJvcihgRXh0ZW5zaW9uIGRpZCBub3QgZXhwb3J0IGFuICdhY3RpdmF0ZScgZnVuY3Rpb24uYClcbiAgICAgICAgICAgIH1cbiAgICAgICAgfSBjYXRjaCAoZXJyKSB7XG4gICAgICAgICAgICBjb25zb2xlLmVycm9yKGVycilcbiAgICAgICAgfVxuICAgIH1cbn1cbiJdfQ==