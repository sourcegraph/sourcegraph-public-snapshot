sub add {
    if (@_ == 2) {
        my $self = shift;
        push(@$self, shift);
        return;
    }
    my($self, %spec) = @_;
    push(@$self, \%spec);
    return;
}

sub something {
    for my $ikey (keys %$item) {
        my $mkey = $ikey;
        if (my $m = $MATCH{$mkey}) {
            my $v = $item->{$ikey};
        }
    }
}
