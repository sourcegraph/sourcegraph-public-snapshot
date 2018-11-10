import { Subject } from 'rxjs';
/** @internal */
export class ExtDocuments {
    constructor(sync) {
        this.sync = sync;
        this.documents = new Map();
        this.textDocumentAdds = new Subject();
        this.onDidOpenTextDocument = this.textDocumentAdds;
    }
    /**
     * Returns the known document with the given URI.
     *
     * @internal
     */
    get(resource) {
        const doc = this.documents.get(resource);
        if (!doc) {
            throw new Error(`document not found: ${resource}`);
        }
        return doc;
    }
    /**
     * If needed, perform a sync with the client to ensure that its pending sends have been received before
     * retrieving this document.
     *
     * @todo This is necessary because hovers can be sent before the document is loaded, and it will cause a
     * "document not found" error.
     */
    async getSync(resource) {
        const doc = this.documents.get(resource);
        if (doc) {
            return doc;
        }
        await this.sync();
        return this.get(resource);
    }
    /**
     * Returns all known documents.
     *
     * @internal
     */
    getAll() {
        return Array.from(this.documents.values());
    }
    $acceptDocumentData(docs) {
        if (!docs) {
            // We don't ever (yet) communicate to the extension when docs are closed.
            return;
        }
        for (const doc of docs) {
            const isNew = !this.documents.has(doc.uri);
            this.documents.set(doc.uri, doc);
            if (isNew) {
                this.textDocumentAdds.next(doc);
            }
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiZG9jdW1lbnRzLmpzIiwic291cmNlUm9vdCI6InNyYy8iLCJzb3VyY2VzIjpbImV4dGVuc2lvbi9hcGkvZG9jdW1lbnRzLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLE9BQU8sRUFBYyxPQUFPLEVBQUUsTUFBTSxNQUFNLENBQUE7QUFTMUMsZ0JBQWdCO0FBQ2hCLE1BQU0sT0FBTyxZQUFZO0lBR3JCLFlBQW9CLElBQXlCO1FBQXpCLFNBQUksR0FBSixJQUFJLENBQXFCO1FBRnJDLGNBQVMsR0FBRyxJQUFJLEdBQUcsRUFBNEIsQ0FBQTtRQTBDL0MscUJBQWdCLEdBQUcsSUFBSSxPQUFPLEVBQWdCLENBQUE7UUFDdEMsMEJBQXFCLEdBQTZCLElBQUksQ0FBQyxnQkFBZ0IsQ0FBQTtJQXpDdkMsQ0FBQztJQUVqRDs7OztPQUlHO0lBQ0ksR0FBRyxDQUFDLFFBQWdCO1FBQ3ZCLE1BQU0sR0FBRyxHQUFHLElBQUksQ0FBQyxTQUFTLENBQUMsR0FBRyxDQUFDLFFBQVEsQ0FBQyxDQUFBO1FBQ3hDLElBQUksQ0FBQyxHQUFHLEVBQUU7WUFDTixNQUFNLElBQUksS0FBSyxDQUFDLHVCQUF1QixRQUFRLEVBQUUsQ0FBQyxDQUFBO1NBQ3JEO1FBQ0QsT0FBTyxHQUFHLENBQUE7SUFDZCxDQUFDO0lBRUQ7Ozs7OztPQU1HO0lBQ0ksS0FBSyxDQUFDLE9BQU8sQ0FBQyxRQUFnQjtRQUNqQyxNQUFNLEdBQUcsR0FBRyxJQUFJLENBQUMsU0FBUyxDQUFDLEdBQUcsQ0FBQyxRQUFRLENBQUMsQ0FBQTtRQUN4QyxJQUFJLEdBQUcsRUFBRTtZQUNMLE9BQU8sR0FBRyxDQUFBO1NBQ2I7UUFDRCxNQUFNLElBQUksQ0FBQyxJQUFJLEVBQUUsQ0FBQTtRQUNqQixPQUFPLElBQUksQ0FBQyxHQUFHLENBQUMsUUFBUSxDQUFDLENBQUE7SUFDN0IsQ0FBQztJQUVEOzs7O09BSUc7SUFDSSxNQUFNO1FBQ1QsT0FBTyxLQUFLLENBQUMsSUFBSSxDQUFDLElBQUksQ0FBQyxTQUFTLENBQUMsTUFBTSxFQUFFLENBQUMsQ0FBQTtJQUM5QyxDQUFDO0lBS00sbUJBQW1CLENBQUMsSUFBK0I7UUFDdEQsSUFBSSxDQUFDLElBQUksRUFBRTtZQUNQLHlFQUF5RTtZQUN6RSxPQUFNO1NBQ1Q7UUFDRCxLQUFLLE1BQU0sR0FBRyxJQUFJLElBQUksRUFBRTtZQUNwQixNQUFNLEtBQUssR0FBRyxDQUFDLElBQUksQ0FBQyxTQUFTLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsQ0FBQTtZQUMxQyxJQUFJLENBQUMsU0FBUyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsR0FBRyxFQUFFLEdBQUcsQ0FBQyxDQUFBO1lBQ2hDLElBQUksS0FBSyxFQUFFO2dCQUNQLElBQUksQ0FBQyxnQkFBZ0IsQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLENBQUE7YUFDbEM7U0FDSjtJQUNMLENBQUM7Q0FDSiIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7IE9ic2VydmFibGUsIFN1YmplY3QgfSBmcm9tICdyeGpzJ1xuaW1wb3J0IHsgVGV4dERvY3VtZW50IH0gZnJvbSAnc291cmNlZ3JhcGgnXG5pbXBvcnQgeyBUZXh0RG9jdW1lbnRJdGVtIH0gZnJvbSAnLi4vLi4vY2xpZW50L3R5cGVzL3RleHREb2N1bWVudCdcblxuLyoqIEBpbnRlcm5hbCAqL1xuZXhwb3J0IGludGVyZmFjZSBFeHREb2N1bWVudHNBUEkge1xuICAgICRhY2NlcHREb2N1bWVudERhdGEoZG9jOiBUZXh0RG9jdW1lbnRJdGVtW10pOiB2b2lkXG59XG5cbi8qKiBAaW50ZXJuYWwgKi9cbmV4cG9ydCBjbGFzcyBFeHREb2N1bWVudHMgaW1wbGVtZW50cyBFeHREb2N1bWVudHNBUEkge1xuICAgIHByaXZhdGUgZG9jdW1lbnRzID0gbmV3IE1hcDxzdHJpbmcsIFRleHREb2N1bWVudEl0ZW0+KClcblxuICAgIGNvbnN0cnVjdG9yKHByaXZhdGUgc3luYzogKCkgPT4gUHJvbWlzZTx2b2lkPikge31cblxuICAgIC8qKlxuICAgICAqIFJldHVybnMgdGhlIGtub3duIGRvY3VtZW50IHdpdGggdGhlIGdpdmVuIFVSSS5cbiAgICAgKlxuICAgICAqIEBpbnRlcm5hbFxuICAgICAqL1xuICAgIHB1YmxpYyBnZXQocmVzb3VyY2U6IHN0cmluZyk6IFRleHREb2N1bWVudCB7XG4gICAgICAgIGNvbnN0IGRvYyA9IHRoaXMuZG9jdW1lbnRzLmdldChyZXNvdXJjZSlcbiAgICAgICAgaWYgKCFkb2MpIHtcbiAgICAgICAgICAgIHRocm93IG5ldyBFcnJvcihgZG9jdW1lbnQgbm90IGZvdW5kOiAke3Jlc291cmNlfWApXG4gICAgICAgIH1cbiAgICAgICAgcmV0dXJuIGRvY1xuICAgIH1cblxuICAgIC8qKlxuICAgICAqIElmIG5lZWRlZCwgcGVyZm9ybSBhIHN5bmMgd2l0aCB0aGUgY2xpZW50IHRvIGVuc3VyZSB0aGF0IGl0cyBwZW5kaW5nIHNlbmRzIGhhdmUgYmVlbiByZWNlaXZlZCBiZWZvcmVcbiAgICAgKiByZXRyaWV2aW5nIHRoaXMgZG9jdW1lbnQuXG4gICAgICpcbiAgICAgKiBAdG9kbyBUaGlzIGlzIG5lY2Vzc2FyeSBiZWNhdXNlIGhvdmVycyBjYW4gYmUgc2VudCBiZWZvcmUgdGhlIGRvY3VtZW50IGlzIGxvYWRlZCwgYW5kIGl0IHdpbGwgY2F1c2UgYVxuICAgICAqIFwiZG9jdW1lbnQgbm90IGZvdW5kXCIgZXJyb3IuXG4gICAgICovXG4gICAgcHVibGljIGFzeW5jIGdldFN5bmMocmVzb3VyY2U6IHN0cmluZyk6IFByb21pc2U8VGV4dERvY3VtZW50PiB7XG4gICAgICAgIGNvbnN0IGRvYyA9IHRoaXMuZG9jdW1lbnRzLmdldChyZXNvdXJjZSlcbiAgICAgICAgaWYgKGRvYykge1xuICAgICAgICAgICAgcmV0dXJuIGRvY1xuICAgICAgICB9XG4gICAgICAgIGF3YWl0IHRoaXMuc3luYygpXG4gICAgICAgIHJldHVybiB0aGlzLmdldChyZXNvdXJjZSlcbiAgICB9XG5cbiAgICAvKipcbiAgICAgKiBSZXR1cm5zIGFsbCBrbm93biBkb2N1bWVudHMuXG4gICAgICpcbiAgICAgKiBAaW50ZXJuYWxcbiAgICAgKi9cbiAgICBwdWJsaWMgZ2V0QWxsKCk6IFRleHREb2N1bWVudFtdIHtcbiAgICAgICAgcmV0dXJuIEFycmF5LmZyb20odGhpcy5kb2N1bWVudHMudmFsdWVzKCkpXG4gICAgfVxuXG4gICAgcHJpdmF0ZSB0ZXh0RG9jdW1lbnRBZGRzID0gbmV3IFN1YmplY3Q8VGV4dERvY3VtZW50PigpXG4gICAgcHVibGljIHJlYWRvbmx5IG9uRGlkT3BlblRleHREb2N1bWVudDogT2JzZXJ2YWJsZTxUZXh0RG9jdW1lbnQ+ID0gdGhpcy50ZXh0RG9jdW1lbnRBZGRzXG5cbiAgICBwdWJsaWMgJGFjY2VwdERvY3VtZW50RGF0YShkb2NzOiBUZXh0RG9jdW1lbnRJdGVtW10gfCBudWxsKTogdm9pZCB7XG4gICAgICAgIGlmICghZG9jcykge1xuICAgICAgICAgICAgLy8gV2UgZG9uJ3QgZXZlciAoeWV0KSBjb21tdW5pY2F0ZSB0byB0aGUgZXh0ZW5zaW9uIHdoZW4gZG9jcyBhcmUgY2xvc2VkLlxuICAgICAgICAgICAgcmV0dXJuXG4gICAgICAgIH1cbiAgICAgICAgZm9yIChjb25zdCBkb2Mgb2YgZG9jcykge1xuICAgICAgICAgICAgY29uc3QgaXNOZXcgPSAhdGhpcy5kb2N1bWVudHMuaGFzKGRvYy51cmkpXG4gICAgICAgICAgICB0aGlzLmRvY3VtZW50cy5zZXQoZG9jLnVyaSwgZG9jKVxuICAgICAgICAgICAgaWYgKGlzTmV3KSB7XG4gICAgICAgICAgICAgICAgdGhpcy50ZXh0RG9jdW1lbnRBZGRzLm5leHQoZG9jKVxuICAgICAgICAgICAgfVxuICAgICAgICB9XG4gICAgfVxufVxuIl19