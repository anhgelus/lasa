LASAD(1)

# NAME

lasad - lasa daemon running the webserver

# SYNOPSIS

*lasad* [*-h*]

*lasad* [*-c* _path/to/config_]

*lasad* run [*-c* _path/to/config_]

*lasad* gen-config [*-c* _path/to/config_]

# DESCRIPTION

*lasad* is the lasa daemon running the webserver.
It generates a RSS or an Atom feed for an ATProto publication.

*lasad* or *lasad run* without any commands starts the daemon.

*lasad gen-config* generates a new config and *overwrites* the current config file.

# OPTIONS

*-c* specifies the path to the config. Default: */etc/lasad.toml*.

*-v* increases verbosity (DEBUG is logged).

*-no-color* disables colors in logs.

*-h* prints the help.

# CONFIGURATION

The configuration file is a TOML file.

*Root section* specifies the _domain_ and the _port_ to listen to.

*cache section* contains information to connect to a _Valkey_ database.
*cache.auth subsection* is used to provide login information.

# SEE ALSO

_lasa_(1)

# AUTHORS

Maintained by William Hergès <@anhgelus.world>.
Up-to-date sources can be found at https://tangled.org/anhgelus.world/lasa/.
