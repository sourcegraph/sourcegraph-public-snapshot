import FluxStoreGroup from "flux/lib/FluxStoreGroup";

import Component from "sourcegraph/Component";

class Container<P, S> extends Component<P, S> {
	_containerSubscriptions: any[];
	_containerStoreGroup: FluxStoreGroup;

	componentDidMount(): void {
		let stores = this.stores();

		let changed = false;
		let setChanged = () => { changed = true; };
		this._containerSubscriptions = stores.map((store) => store.addListener(setChanged));

		this._containerStoreGroup = new FluxStoreGroup(stores, () => {
			if (changed) {
				this.setState({} as S);
			}
			changed = false;
		});
	}

	componentWillUnmount(): void {
		if (this._containerStoreGroup) { this._containerStoreGroup.release(); }
		if (this._containerSubscriptions) {
			this._containerSubscriptions.forEach((subscription) => { subscription.remove(); });
		}
	}

	stores(): FluxUtils.Store<any>[] {
		return [];
	}
}

export default Container;
