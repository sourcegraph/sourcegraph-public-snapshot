#[derive(Debug, PartialEq, Eq, Default)]
pub struct PackedRange {
    pub start_line: i32,
    pub start_col: i32,
    pub end_line: i32,
    pub end_col: i32,
}

impl PackedRange {
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

    /// Checks if the range is equal to the given vector.
    /// If the other vector is not a valid PackedRange then it returns false
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
}

impl PartialOrd for PackedRange {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        (self.start_line, self.end_line, self.start_col).partial_cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}

impl Ord for PackedRange {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        (self.start_line, self.end_line, self.start_col).cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}
