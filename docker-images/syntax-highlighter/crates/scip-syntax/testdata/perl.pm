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
