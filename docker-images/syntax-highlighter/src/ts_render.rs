use std::io::BufRead;
use tree_sitter::LossyUtf8;
use tree_sitter_highlight::Error;
use tree_sitter_highlight::{Highlight, HighlightEvent};

const BUFFER_HTML_RESERVE_CAPACITY: usize = 10 * 1024;

/// Represents the reason why syntax highlighting failed.

/// Converts a general-purpose syntax highlighting iterator into a sequence of lines of HTML.
pub struct TableHtmlRenderer {
    pub highlighted: Vec<u8>,
    pub html: Vec<u8>,
}

/// Our version of `tree_sitter_highlight::HtmlRenderer`, which emits stuff as a table.
///
/// You can see the original version in the tree_sitter_highlight crate.
impl TableHtmlRenderer {
    pub fn new() -> Self {
        TableHtmlRenderer {
            // TODO: This is just wasting space, but now sure how to stream this and correctly
            // handle line endings. Instead I wait til things are done, and then split the lines
            // afterwards.
            //
            // Could perhaps match on something or modify the Highlighter, but it seems very hard.
            // This way is simple but does save the string twice basically.
            highlighted: Vec::with_capacity(BUFFER_HTML_RESERVE_CAPACITY),
            html: Vec::new(),
        }
    }

    pub fn render<'a, F>(
        &mut self,
        highlighter: impl Iterator<Item = Result<HighlightEvent, Error>>,
        source: &'a [u8],
        attribute_callback: &F,
    ) -> Result<(), Error>
    where
        F: Fn(Highlight) -> &'a [u8],
    {
        let mut highlights = Vec::new();
        for event in highlighter {
            match event {
                Ok(HighlightEvent::HighlightStart(s)) => {
                    highlights.push(s);
                    self.start_highlight(s, attribute_callback);
                }
                Ok(HighlightEvent::HighlightEnd) => {
                    highlights.pop();
                    self.end_highlight();
                }
                Ok(HighlightEvent::Source { start, end }) => {
                    self.add_text(&source[start..end], &highlights, attribute_callback);
                }
                Err(a) => return Err(a),
            }
        }

        if self.highlighted.last() != Some(&b'\n') {
            self.highlighted.push(b'\n');
        }

        // Just guess that we need something twice as long, so we don't have a lot of resizes
        self.html = Vec::with_capacity(self.highlighted.len() * 2);

        // This is the same format as ClassedTableGenerator
        //
        // TODO: Could probably try and make these share some code :)
        //
        //     <tr>
        //       <td class="line" data-line="1">
        //       <td class="code">
        //         <span class="hl-source hl-go">
        //           <span class="hl-keyword hl-control hl-go">package</span>
        //           main
        //         </span>
        //       </td>
        //     </tr>
        self.html.extend_from_slice("<table><tbody>".as_bytes());
        for (idx, line) in self.highlighted.lines().enumerate() {
            let line = line.unwrap();
            self.html.extend_from_slice(
                format!(
                    r#"<tr><td class="line" data-line="{}"><td class="code"><div>{}</div></td></tr>"#,
                    idx + 1,
                    line
                )
                .as_bytes(),
            );
        }
        self.html.extend_from_slice("</tbody></table>".as_bytes());

        Ok(())
    }

    fn start_highlight<'a, F>(&mut self, h: Highlight, attribute_callback: &F)
    where
        F: Fn(Highlight) -> &'a [u8],
    {
        let attribute_string = (attribute_callback)(h);
        self.highlighted.extend(b"<span");
        if !attribute_string.is_empty() {
            self.highlighted.extend(b" ");
            self.highlighted.extend(attribute_string);
        }
        self.highlighted.extend(b">");
    }

    fn end_highlight(&mut self) {
        self.highlighted.extend(b"</span>");
    }

    fn add_text<'a, F>(&mut self, src: &[u8], highlights: &Vec<Highlight>, attribute_callback: &F)
    where
        F: Fn(Highlight) -> &'a [u8],
    {
        let mut last_char_was_cr = false;
        for c in LossyUtf8::new(src).flat_map(|p| p.bytes()) {
            // Don't render carriage return characters, but allow lone carriage returns (not
            // followed by line feeds) to be styled via the attribute callback.
            if c == b'\r' {
                last_char_was_cr = true;
                continue;
            }
            if last_char_was_cr {
                last_char_was_cr = false;
            }

            // At line boundaries, close and re-open all of the open tags.
            if c == b'\n' {
                highlights.iter().for_each(|_| self.end_highlight());
                self.highlighted.push(c);
                highlights
                    .iter()
                    .for_each(|scope| self.start_highlight(*scope, attribute_callback));
            } else if let Some(escape) = html_escape(c) {
                self.highlighted.extend_from_slice(escape);
            } else {
                self.highlighted.push(c);
            }
        }
    }
}

pub fn html_escape(c: u8) -> Option<&'static [u8]> {
    match c as char {
        '>' => Some(b"&gt;"),
        '<' => Some(b"&lt;"),
        '&' => Some(b"&amp;"),
        '\'' => Some(b"&#39;"),
        '"' => Some(b"&quot;"),
        _ => None,
    }
}
