#!/bin/sh

PLATFORM=$(uname)

case "$PLATFORM" in
	FreeBSD)
		install -v gruptime/gruptime /usr/local/bin
		install -v rc.d/gruptime /usr/local/etc/rc.d
		;;
	Linux)
		install -v gruptime/gruptime /usr/local/bin
		install -v rc.d/gruptime.service /usr/lib/systemd/system
		;;
	SunOS)
		install -v gruptime/gruptime /usr/bin
		echo "no smf service file yet"
		;;
	*)
		echo "no automatic install for $PLATFORM"
		;;
esac

echo "be sure to make a /etc/gruptime.conf or /usr/local/etc/gruptime.conf"

