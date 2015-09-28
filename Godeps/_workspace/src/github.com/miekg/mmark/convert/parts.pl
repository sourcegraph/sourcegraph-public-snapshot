#!/usr/bin/perl

# Quick and dirty perl script to make pandoc2rfc markdown fully
# compliant with mmark markdown. Fixes figure and table captions,
# makes RFC references proper citations and uses the new index syntax.
#
# BUGS: makes all references normative
# Does not handle I-D references.
#
# pandoc --atx-headers middle.mkd -t markdown_phpextra | ./part.pl | tee part.md

@doc = <>;

# Fix figures
#
# Look for code blocks, mark them and see if there is an footnote (=caption)
# after them
#
# Input looks like:
#
# The RDATA of the NEXT RR is as shown below.
#
#                          1 1 1 1 1 1 1 1 1 1 2 2 2 2 2 2 2 2 2 2 3 3
#      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
#     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
#     /                         Type Bit Maps                         /
#     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
#
# [^2]

# %footnote holds the information:
# key is the footnote ID and a starting line for the code block or table, see we can
# insert the reference. We also keep track when the figure ends, because that
# is where we need to insert the code.
# $footnote{"1"} -> [type, startline, stopline, reference and caption text]
# where type is "Figure: '  or "Table: "

for ( $i = 0; $i < scalar @doc; $i++) {
    if ( $doc[$i] =~ /^$/ ) {
        # if the next line starts with 4 spaces - it's a code block
        if ( $doc[$i+1] =~ /^    / ) {
            $blockStart = $i;
            $i++;
            while ( $doc[$i] =~ /^    / ) {
                    $i++;
            }
            # code blocks has stopped, this must be an empty line, the *next* line after
            # that *may* have an footnote
            $blockEnd = $i-1;
            $i++;
            if ( $doc[$i] =~ /^\[\^(\d+)\]/ ) {
                $doc[$i] = "";
                $note = $1;
                $footnote{$note} = [ "Figure: ", $blockStart, $blockEnd ];
            }
        }
    }
    # footnote texts
    # caption:   [^2]: fig:next-wire::The NEXT on-the-wire format.
    # index:     [^1]: ^item^ subitem
    if ( $doc[$i] =~ /^\[\^(\d)+\]\:/ ) {
        $note = $1;
        # caption tag until an empty line
        $foot = "";
        while ( $doc[$i] !~ /^$/ ) {
            $foot = $foot . $doc[$i];
            # clear text
            $doc[$i] = "";
            $i++;
        }
        # these footnotes come last so we already havve $footnote set.
        push @{$footnote{$note}}, $foot;
    }
}

# the above loop takes care of the code blocks caption, the ones that are left *must* be
# table caption. Loop through it can and perform the same trick, but this for tables.
for ( $i = 0; $i < scalar @doc; $i++) {
    if ( $doc[$i] =~ /^$/ ) {
        # if the next line starts with a pipe - it's a code table
        if ( $doc[$i+1] =~ /^\|/ ) {
            $blockStart = $i;
            $i++;
            while ( $doc[$i] =~ /^\|/ ) {
                    $i++;
            }
            # table blocks has stopped, this must be an empty line, the *next* line after
            # that *may* have an footnote
            $blockEnd = $i-1;
            $i++;
            if ( $doc[$i] =~ /^\[\^(\d+)\]/ ) {
                $doc[$i] = "";
                $note = $1;
                $notetext = @{$footnote{$note}}[0];
                $footnote{$note} = [ "Table: ", $blockStart, $blockEnd, $notetext ];
            }
        }
    }
}

foreach $k (keys %footnote) {
    $type =  ${$footnote{$k}}[0];
    $begin =  ${$footnote{$k}}[1];
    $end =  ${$footnote{$k}}[2];
    $text = ${$footnote{$k}}[3];
    if ( $begin > 0 ) {
        ($anchor, $caption) = split /::/, $text;
        # strip anchor of the footnote prefix
        $anchor =~ s/^\[\^\d+\]: //;
        # now begin will get the reference
        if ($anchor ne "") {
            $doc[$begin] = $doc[$begin] . "{#$anchor}\n\n";
        }
        # caption can not be empty
        $doc[$end] = $doc[$end] . $type . $caption;
        delete $footnote{$k};
    }
}

sub foot2index(@) {
    # [^1]: <sup>item</sup> subitem -> (((item, subtitem))
    if ( $_[0] =~ m|<sup>(.*)</sup>(.*)| ) {
        return "((($1,$2)))";
    }
    $_[0] =~ s/^\[\^\d+\]\: //;
    return $_[0];
}

foreach (@doc) {
    # [](#RFC5155) -> [@!RFC5155]
    s/\[\]\(\#RFC(\d+)\)/[@!RFC\1]/g;
    # any footnotes left are indices
    s/\[\^(\d+)\]/ foot2index @{$footnote{$1}} /eg;
    print;
}
