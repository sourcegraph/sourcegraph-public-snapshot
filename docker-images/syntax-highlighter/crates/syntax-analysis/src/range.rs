use tree_sitter::Node;

#[derive(Debug, PartialEq, Eq, Default, Hash, Copy, Clone)]
pub struct Range {
    pub start_line: i32,
    pub start_col: i32,
    pub end_line: i32,
    pub end_col: i32,
}

impl Range {
    // TODO: I don't know how much I love just returning an error here...
    //       Maybe we should make an infallible version too. It's just a bit annoying
    //       to pack and unpack all of these for effectively no reason.
    //       But for now it's good.
    pub fn from_vec(v: &[i32]) -> Option<Self> {
        match v.len() {
            3 => Some(Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[0],
                end_col: v[2],
            }),
            4 => Some(Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[2],
                end_col: v[3],
            }),
            _ => None,
        }
    }

    pub fn to_vec(&self) -> Vec<i32> {
        if self.start_line == self.end_line {
            vec![self.start_line, self.start_col, self.end_col]
        } else {
            vec![self.start_line, self.start_col, self.end_line, self.end_col]
        }
    }

    /// Checks if the range is equal to the given vector.
    /// If the other vector is not a valid Range then it returns false
    pub fn eq_vec(&self, v: &[i32]) -> bool {
        match v.len() {
            3 => {
                self.start_line == v[0]
                    && self.start_col == v[1]
                    && self.end_line == v[0]
                    && self.end_col == v[2]
            }
            4 => {
                self.start_line == v[0]
                    && self.start_col == v[1]
                    && self.end_line == v[2]
                    && self.end_col == v[3]
            }
            _ => false,
        }
    }

    pub fn contains(&self, other: &Range) -> bool {
        other.start_line >= self.start_line
            && other.end_line <= self.end_line
            && (other.start_line != self.start_line || other.start_col >= self.start_col)
            && (other.end_line != self.end_line || other.end_col <= self.end_col)
    }
}

impl PartialOrd for Range {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for Range {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        (self.start_line, self.end_line, self.start_col).cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}

impl<'a> From<Node<'a>> for Range {
    fn from(node: Node<'a>) -> Self {
        let start = node.start_position();
        let end = node.end_position();
        Self {
            start_line: start.row as i32,
            start_col: start.column as i32,
            end_line: end.row as i32,
            end_col: end.column as i32,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::Range;

    #[test]
    fn test_packed_range_contains() {
        let outer = Range {
            start_line: 1,
            start_col: 1,
            end_line: 4,
            end_col: 4,
        };

        let inner = Range {
            start_line: 2,
            start_col: 2,
            end_line: 3,
            end_col: 3,
        };

        let overlapping = Range {
            start_line: 3,
            start_col: 3,
            end_line: 5,
            end_col: 5,
        };

        let outside = Range {
            start_line: 5,
            start_col: 5,
            end_line: 6,
            end_col: 6,
        };

        let same = Range {
            start_line: 1,
            start_col: 1,
            end_line: 4,
            end_col: 4,
        };

        assert!(outer.contains(&inner));
        assert!(!outer.contains(&overlapping));
        assert!(!outer.contains(&outside));
        assert!(outer.contains(&same));
    }
}
