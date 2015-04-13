#!/bin/bash -e

#executed when the container starts
echo Setting credentials to $USERNAME:$PASSWORD

PASSWORD=$(perl -e 'print crypt($ARGV[0], "password")' $PASSWORD)
useradd --shell /bin/sh --create-home --password $PASSWORD $USERNAME
chown -R $USERNAME:$USERNAME /ftp

#start ftp server
service proftpd start

#start the watcher
fswatcher /ftp