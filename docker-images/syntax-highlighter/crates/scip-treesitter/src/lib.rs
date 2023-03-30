use tree_sitter::Node;

pub mod snapshot;
pub mod types;

pub mod prelude {
    pub use super::{ContainsNode, NodeToScipRange};
}

pub trait ContainsNode {
    fn contains_node(&self, node: &Node) -> bool;
}

impl<'a> ContainsNode for Node<'a> {
    fn contains_node(&self, node: &Node) -> bool {
        self.start_byte() <= node.start_byte() && self.end_byte() >= node.end_byte()
    }
}

pub trait NodeToScipRange {
    fn to_scip_range(&self) -> Vec<i32>;
}

impl<'a> NodeToScipRange for Node<'a> {
    fn to_scip_range(&self) -> Vec<i32> {
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

// #[allow(dead_code)]
// pub fn walk_child(node: &Node, depth: usize) {
//     let mut cursor = node.walk();
//
//     node.children(&mut cursor).for_each(|child| {
//         println!(
//             "{}{:?} {} {}",
//             " ".repeat(depth),
//             child,
//             child.start_position().column,
//             child.end_position().column
//         );
//
//         walk_child(&child, depth + 1);
//     });
// }
