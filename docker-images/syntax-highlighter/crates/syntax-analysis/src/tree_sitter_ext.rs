use tree_sitter::Node;

/// Extension methods for Tree-sitter's `Node` type.
pub trait NodeExt {
    fn scip_range(&self) -> Vec<i32>;
}

impl<'a> NodeExt for Node<'a> {
    fn scip_range(&self) -> Vec<i32> {
        let start_position = self.start_position();
        let end_position = self.end_position();

        if start_position.row == end_position.row {
            vec![
                start_position.row as i32,
                start_position.column as i32,
                end_position.column as i32,
            ]
        } else {
            vec![
                start_position.row as i32,
                start_position.column as i32,
                end_position.row as i32,
                end_position.column as i32,
            ]
        }
    }
}
