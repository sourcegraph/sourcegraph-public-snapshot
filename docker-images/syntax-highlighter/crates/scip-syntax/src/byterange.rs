#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct ByteRange {
    pub start: usize,
    pub end: usize,
}

impl ByteRange {
    pub fn contains(&self, other: &Self) -> bool {
        self.start <= other.start && self.end >= other.end
    }
}
