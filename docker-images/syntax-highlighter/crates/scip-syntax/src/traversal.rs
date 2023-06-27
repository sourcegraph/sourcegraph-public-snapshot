use scip_treesitter::types::PackedRange;

pub trait RangeContainer: Sized {
    fn range(&self) -> &PackedRange;
}

pub trait ChildContainer: RangeContainer {
    fn children(&mut self) -> &mut Vec<Self>;

    fn find_container<'a>(&'a mut self, item: &impl RangeContainer) -> Option<&'a mut Self> {
        self.children()
            .iter_mut()
            .find(|child| child.range().contains(item.range()))
    }

    fn insert_item<'a, T: RangeContainer>(
        &'a mut self,
        item: T,
        push_item: &mut dyn FnMut(&mut Self, T),
    ) {
        if let Some(child) = self.find_container(&item) {
            child.insert_item(item, push_item)
        } else {
            push_item(self, item)
        }
    }
}
