import { Subscription } from 'rxjs';
import { handleRequests } from '../../common/proxy';
/** @internal */
export class ClientContext {
    constructor(connection, updateContext) {
        this.updateContext = updateContext;
        this.subscriptions = new Subscription();
        /**
         * Context keys set by this server. To ensure that context values are cleaned up, all context properties that
         * the server set are cleared upon deinitialization. This errs on the side of clearing too much (if another
         * server set the same context keys after this server, then those keys would also be cleared when this server's
         * client deinitializes).
         */
        this.keys = new Set();
        handleRequests(connection, 'context', this);
    }
    $acceptContextUpdates(updates) {
        for (const key of Object.keys(updates)) {
            this.keys.add(key);
        }
        this.updateContext(updates);
    }
    unsubscribe() {
        /**
         * Clear all context properties whose keys were ever set by the server. See {@link ClientContext#keys}.
         */
        const updates = {};
        for (const key of this.keys) {
            updates[key] = null;
        }
        this.keys.clear();
        this.updateContext(updates);
        this.subscriptions.unsubscribe();
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY29udGV4dC5qcyIsInNvdXJjZVJvb3QiOiJzcmMvIiwic291cmNlcyI6WyJjbGllbnQvYXBpL2NvbnRleHQudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsT0FBTyxFQUFFLFlBQVksRUFBRSxNQUFNLE1BQU0sQ0FBQTtBQUVuQyxPQUFPLEVBQUUsY0FBYyxFQUFFLE1BQU0sb0JBQW9CLENBQUE7QUFRbkQsZ0JBQWdCO0FBQ2hCLE1BQU0sT0FBTyxhQUFhO0lBV3RCLFlBQVksVUFBc0IsRUFBVSxhQUErQztRQUEvQyxrQkFBYSxHQUFiLGFBQWEsQ0FBa0M7UUFWbkYsa0JBQWEsR0FBRyxJQUFJLFlBQVksRUFBRSxDQUFBO1FBRTFDOzs7OztXQUtHO1FBQ0ssU0FBSSxHQUFHLElBQUksR0FBRyxFQUFVLENBQUE7UUFHNUIsY0FBYyxDQUFDLFVBQVUsRUFBRSxTQUFTLEVBQUUsSUFBSSxDQUFDLENBQUE7SUFDL0MsQ0FBQztJQUVNLHFCQUFxQixDQUFDLE9BQXNCO1FBQy9DLEtBQUssTUFBTSxHQUFHLElBQUksTUFBTSxDQUFDLElBQUksQ0FBQyxPQUFPLENBQUMsRUFBRTtZQUNwQyxJQUFJLENBQUMsSUFBSSxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsQ0FBQTtTQUNyQjtRQUNELElBQUksQ0FBQyxhQUFhLENBQUMsT0FBTyxDQUFDLENBQUE7SUFDL0IsQ0FBQztJQUVNLFdBQVc7UUFDZDs7V0FFRztRQUNILE1BQU0sT0FBTyxHQUFrQixFQUFFLENBQUE7UUFDakMsS0FBSyxNQUFNLEdBQUcsSUFBSSxJQUFJLENBQUMsSUFBSSxFQUFFO1lBQ3pCLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxJQUFJLENBQUE7U0FDdEI7UUFDRCxJQUFJLENBQUMsSUFBSSxDQUFDLEtBQUssRUFBRSxDQUFBO1FBQ2pCLElBQUksQ0FBQyxhQUFhLENBQUMsT0FBTyxDQUFDLENBQUE7UUFFM0IsSUFBSSxDQUFDLGFBQWEsQ0FBQyxXQUFXLEVBQUUsQ0FBQTtJQUNwQyxDQUFDO0NBQ0oiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBTdWJzY3JpcHRpb24gfSBmcm9tICdyeGpzJ1xuaW1wb3J0IHsgQ29udGV4dFZhbHVlcyB9IGZyb20gJ3NvdXJjZWdyYXBoJ1xuaW1wb3J0IHsgaGFuZGxlUmVxdWVzdHMgfSBmcm9tICcuLi8uLi9jb21tb24vcHJveHknXG5pbXBvcnQgeyBDb25uZWN0aW9uIH0gZnJvbSAnLi4vLi4vcHJvdG9jb2wvanNvbnJwYzIvY29ubmVjdGlvbidcblxuLyoqIEBpbnRlcm5hbCAqL1xuZXhwb3J0IGludGVyZmFjZSBDbGllbnRDb250ZXh0QVBJIHtcbiAgICAkYWNjZXB0Q29udGV4dFVwZGF0ZXModXBkYXRlczogQ29udGV4dFZhbHVlcyk6IHZvaWRcbn1cblxuLyoqIEBpbnRlcm5hbCAqL1xuZXhwb3J0IGNsYXNzIENsaWVudENvbnRleHQgaW1wbGVtZW50cyBDbGllbnRDb250ZXh0QVBJIHtcbiAgICBwcml2YXRlIHN1YnNjcmlwdGlvbnMgPSBuZXcgU3Vic2NyaXB0aW9uKClcblxuICAgIC8qKlxuICAgICAqIENvbnRleHQga2V5cyBzZXQgYnkgdGhpcyBzZXJ2ZXIuIFRvIGVuc3VyZSB0aGF0IGNvbnRleHQgdmFsdWVzIGFyZSBjbGVhbmVkIHVwLCBhbGwgY29udGV4dCBwcm9wZXJ0aWVzIHRoYXRcbiAgICAgKiB0aGUgc2VydmVyIHNldCBhcmUgY2xlYXJlZCB1cG9uIGRlaW5pdGlhbGl6YXRpb24uIFRoaXMgZXJycyBvbiB0aGUgc2lkZSBvZiBjbGVhcmluZyB0b28gbXVjaCAoaWYgYW5vdGhlclxuICAgICAqIHNlcnZlciBzZXQgdGhlIHNhbWUgY29udGV4dCBrZXlzIGFmdGVyIHRoaXMgc2VydmVyLCB0aGVuIHRob3NlIGtleXMgd291bGQgYWxzbyBiZSBjbGVhcmVkIHdoZW4gdGhpcyBzZXJ2ZXInc1xuICAgICAqIGNsaWVudCBkZWluaXRpYWxpemVzKS5cbiAgICAgKi9cbiAgICBwcml2YXRlIGtleXMgPSBuZXcgU2V0PHN0cmluZz4oKVxuXG4gICAgY29uc3RydWN0b3IoY29ubmVjdGlvbjogQ29ubmVjdGlvbiwgcHJpdmF0ZSB1cGRhdGVDb250ZXh0OiAodXBkYXRlczogQ29udGV4dFZhbHVlcykgPT4gdm9pZCkge1xuICAgICAgICBoYW5kbGVSZXF1ZXN0cyhjb25uZWN0aW9uLCAnY29udGV4dCcsIHRoaXMpXG4gICAgfVxuXG4gICAgcHVibGljICRhY2NlcHRDb250ZXh0VXBkYXRlcyh1cGRhdGVzOiBDb250ZXh0VmFsdWVzKTogdm9pZCB7XG4gICAgICAgIGZvciAoY29uc3Qga2V5IG9mIE9iamVjdC5rZXlzKHVwZGF0ZXMpKSB7XG4gICAgICAgICAgICB0aGlzLmtleXMuYWRkKGtleSlcbiAgICAgICAgfVxuICAgICAgICB0aGlzLnVwZGF0ZUNvbnRleHQodXBkYXRlcylcbiAgICB9XG5cbiAgICBwdWJsaWMgdW5zdWJzY3JpYmUoKTogdm9pZCB7XG4gICAgICAgIC8qKlxuICAgICAgICAgKiBDbGVhciBhbGwgY29udGV4dCBwcm9wZXJ0aWVzIHdob3NlIGtleXMgd2VyZSBldmVyIHNldCBieSB0aGUgc2VydmVyLiBTZWUge0BsaW5rIENsaWVudENvbnRleHQja2V5c30uXG4gICAgICAgICAqL1xuICAgICAgICBjb25zdCB1cGRhdGVzOiBDb250ZXh0VmFsdWVzID0ge31cbiAgICAgICAgZm9yIChjb25zdCBrZXkgb2YgdGhpcy5rZXlzKSB7XG4gICAgICAgICAgICB1cGRhdGVzW2tleV0gPSBudWxsXG4gICAgICAgIH1cbiAgICAgICAgdGhpcy5rZXlzLmNsZWFyKClcbiAgICAgICAgdGhpcy51cGRhdGVDb250ZXh0KHVwZGF0ZXMpXG5cbiAgICAgICAgdGhpcy5zdWJzY3JpcHRpb25zLnVuc3Vic2NyaWJlKClcbiAgICB9XG59XG4iXX0=