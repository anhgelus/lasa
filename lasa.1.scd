LASA(1)

# NAME

lasa - display Standard.site publications

# SYNOPSIS

*lasa* [*-h*] publication _identifier_ [_rkey_]

*lasa* [*-h*] rss _identifier_ _rkey_

*lasa* [*-h*] atom _identifier_ _rkey_

# DESCRIPTION

*lasa publication* displays information for a Standard.site publication.

*lasa rss* renders the RSS feed for the given publication.

*lasa atom* renders the Atom feed for the given publication.

# OPTIONS

_identifier_ is the identifier of an ATProto account.
It can an *Handle* or a *DID*.

_rkey_ is the record key of the publication.

*-h* prints the help.

# SEE ALSO

_lasad_(1)

# AUTHORS

Maintained by William Hergès <@anhgelus.world>.
Up-to-date sources can be found at https://tangled.org/anhgelus.world/lasa/.
