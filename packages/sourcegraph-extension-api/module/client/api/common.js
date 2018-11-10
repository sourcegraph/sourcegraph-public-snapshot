/**
 * Manages a map of subscriptions keyed on a numeric ID.
 *
 * @internal
 */
export class SubscriptionMap {
    constructor() {
        this.map = new Map();
    }
    /**
     * Adds a new subscription with the given {@link id}.
     *
     * @param id - A unique identifier for this subscription among all other entries in this map.
     * @param subscription - The subscription, unsubscribed when {@link SubscriptionMap#remove} is called.
     * @throws If there already exists an entry with the given {@link id}.
     */
    add(id, subscription) {
        if (this.map.has(id)) {
            throw new Error(`subscription already exists with ID ${id}`);
        }
        this.map.set(id, subscription);
    }
    /**
     * Unsubscribes the subscription that was previously added with the given {@link id}, and removes it from the
     * map.
     */
    remove(id) {
        const subscription = this.map.get(id);
        if (!subscription) {
            throw new Error(`no subscription with ID ${id}`);
        }
        try {
            subscription.unsubscribe();
        }
        finally {
            this.map.delete(id);
        }
    }
    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    unsubscribe() {
        try {
            for (const subscription of this.map.values()) {
                subscription.unsubscribe();
            }
        }
        finally {
            this.map.clear();
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY29tbW9uLmpzIiwic291cmNlUm9vdCI6InNyYy8iLCJzb3VyY2VzIjpbImNsaWVudC9hcGkvY29tbW9uLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUVBOzs7O0dBSUc7QUFDSCxNQUFNLE9BQU8sZUFBZTtJQUE1QjtRQUNZLFFBQUcsR0FBRyxJQUFJLEdBQUcsRUFBMEIsQ0FBQTtJQTRDbkQsQ0FBQztJQTFDRzs7Ozs7O09BTUc7SUFDSSxHQUFHLENBQUMsRUFBVSxFQUFFLFlBQTRCO1FBQy9DLElBQUksSUFBSSxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsRUFBRSxDQUFDLEVBQUU7WUFDbEIsTUFBTSxJQUFJLEtBQUssQ0FBQyx1Q0FBdUMsRUFBRSxFQUFFLENBQUMsQ0FBQTtTQUMvRDtRQUNELElBQUksQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLEVBQUUsRUFBRSxZQUFZLENBQUMsQ0FBQTtJQUNsQyxDQUFDO0lBRUQ7OztPQUdHO0lBQ0ksTUFBTSxDQUFDLEVBQVU7UUFDcEIsTUFBTSxZQUFZLEdBQUcsSUFBSSxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsRUFBRSxDQUFDLENBQUE7UUFDckMsSUFBSSxDQUFDLFlBQVksRUFBRTtZQUNmLE1BQU0sSUFBSSxLQUFLLENBQUMsMkJBQTJCLEVBQUUsRUFBRSxDQUFDLENBQUE7U0FDbkQ7UUFDRCxJQUFJO1lBQ0EsWUFBWSxDQUFDLFdBQVcsRUFBRSxDQUFBO1NBQzdCO2dCQUFTO1lBQ04sSUFBSSxDQUFDLEdBQUcsQ0FBQyxNQUFNLENBQUMsRUFBRSxDQUFDLENBQUE7U0FDdEI7SUFDTCxDQUFDO0lBRUQ7O09BRUc7SUFDSSxXQUFXO1FBQ2QsSUFBSTtZQUNBLEtBQUssTUFBTSxZQUFZLElBQUksSUFBSSxDQUFDLEdBQUcsQ0FBQyxNQUFNLEVBQUUsRUFBRTtnQkFDMUMsWUFBWSxDQUFDLFdBQVcsRUFBRSxDQUFBO2FBQzdCO1NBQ0o7Z0JBQVM7WUFDTixJQUFJLENBQUMsR0FBRyxDQUFDLEtBQUssRUFBRSxDQUFBO1NBQ25CO0lBQ0wsQ0FBQztDQUNKIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgVW5zdWJzY3JpYmFibGUgfSBmcm9tICdzb3VyY2VncmFwaCdcblxuLyoqXG4gKiBNYW5hZ2VzIGEgbWFwIG9mIHN1YnNjcmlwdGlvbnMga2V5ZWQgb24gYSBudW1lcmljIElELlxuICpcbiAqIEBpbnRlcm5hbFxuICovXG5leHBvcnQgY2xhc3MgU3Vic2NyaXB0aW9uTWFwIHtcbiAgICBwcml2YXRlIG1hcCA9IG5ldyBNYXA8bnVtYmVyLCBVbnN1YnNjcmliYWJsZT4oKVxuXG4gICAgLyoqXG4gICAgICogQWRkcyBhIG5ldyBzdWJzY3JpcHRpb24gd2l0aCB0aGUgZ2l2ZW4ge0BsaW5rIGlkfS5cbiAgICAgKlxuICAgICAqIEBwYXJhbSBpZCAtIEEgdW5pcXVlIGlkZW50aWZpZXIgZm9yIHRoaXMgc3Vic2NyaXB0aW9uIGFtb25nIGFsbCBvdGhlciBlbnRyaWVzIGluIHRoaXMgbWFwLlxuICAgICAqIEBwYXJhbSBzdWJzY3JpcHRpb24gLSBUaGUgc3Vic2NyaXB0aW9uLCB1bnN1YnNjcmliZWQgd2hlbiB7QGxpbmsgU3Vic2NyaXB0aW9uTWFwI3JlbW92ZX0gaXMgY2FsbGVkLlxuICAgICAqIEB0aHJvd3MgSWYgdGhlcmUgYWxyZWFkeSBleGlzdHMgYW4gZW50cnkgd2l0aCB0aGUgZ2l2ZW4ge0BsaW5rIGlkfS5cbiAgICAgKi9cbiAgICBwdWJsaWMgYWRkKGlkOiBudW1iZXIsIHN1YnNjcmlwdGlvbjogVW5zdWJzY3JpYmFibGUpOiB2b2lkIHtcbiAgICAgICAgaWYgKHRoaXMubWFwLmhhcyhpZCkpIHtcbiAgICAgICAgICAgIHRocm93IG5ldyBFcnJvcihgc3Vic2NyaXB0aW9uIGFscmVhZHkgZXhpc3RzIHdpdGggSUQgJHtpZH1gKVxuICAgICAgICB9XG4gICAgICAgIHRoaXMubWFwLnNldChpZCwgc3Vic2NyaXB0aW9uKVxuICAgIH1cblxuICAgIC8qKlxuICAgICAqIFVuc3Vic2NyaWJlcyB0aGUgc3Vic2NyaXB0aW9uIHRoYXQgd2FzIHByZXZpb3VzbHkgYWRkZWQgd2l0aCB0aGUgZ2l2ZW4ge0BsaW5rIGlkfSwgYW5kIHJlbW92ZXMgaXQgZnJvbSB0aGVcbiAgICAgKiBtYXAuXG4gICAgICovXG4gICAgcHVibGljIHJlbW92ZShpZDogbnVtYmVyKTogdm9pZCB7XG4gICAgICAgIGNvbnN0IHN1YnNjcmlwdGlvbiA9IHRoaXMubWFwLmdldChpZClcbiAgICAgICAgaWYgKCFzdWJzY3JpcHRpb24pIHtcbiAgICAgICAgICAgIHRocm93IG5ldyBFcnJvcihgbm8gc3Vic2NyaXB0aW9uIHdpdGggSUQgJHtpZH1gKVxuICAgICAgICB9XG4gICAgICAgIHRyeSB7XG4gICAgICAgICAgICBzdWJzY3JpcHRpb24udW5zdWJzY3JpYmUoKVxuICAgICAgICB9IGZpbmFsbHkge1xuICAgICAgICAgICAgdGhpcy5tYXAuZGVsZXRlKGlkKVxuICAgICAgICB9XG4gICAgfVxuXG4gICAgLyoqXG4gICAgICogVW5zdWJzY3JpYmVzIGFsbCBzdWJzY3JpcHRpb25zIGluIHRoaXMgbWFwIGFuZCBjbGVhcnMgaXQuXG4gICAgICovXG4gICAgcHVibGljIHVuc3Vic2NyaWJlKCk6IHZvaWQge1xuICAgICAgICB0cnkge1xuICAgICAgICAgICAgZm9yIChjb25zdCBzdWJzY3JpcHRpb24gb2YgdGhpcy5tYXAudmFsdWVzKCkpIHtcbiAgICAgICAgICAgICAgICBzdWJzY3JpcHRpb24udW5zdWJzY3JpYmUoKVxuICAgICAgICAgICAgfVxuICAgICAgICB9IGZpbmFsbHkge1xuICAgICAgICAgICAgdGhpcy5tYXAuY2xlYXIoKVxuICAgICAgICB9XG4gICAgfVxufVxuIl19