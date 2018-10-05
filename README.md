Renamer
========

This CLI application renames your show with the episode title.

Rename your show with format "Show Name - SxxExx - Episode Name"

## Usage


    renamer [-h|--help] -p|--path "\<value>" [-a|--auto] [-t|--test] [-r|--regexp "\<value>"]

## Arguments:

    -h  --help    Print help information
    -p  --path    Path to folder to scan.
    -a  --auto    Automatically rename your show if set.
    -t  --test    Do a test run without renaming anything.
    -r  --regexp  /!\ EXPERIMENTAL /!\ Replace the current regexp, it ABSOLUTELY
                  needs the following capture groups : name, season and episode.