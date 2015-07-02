tunnelhub
=========

Unsafe TCP tunneling tool for massive control.

Here are 3 components: `hub` runs on the edge server, `sub` runs on the slave server, and you use `proxy` to connect to slaves.

Usage
=====

hub:

    ./hub

sub:

    ./sub -server="example.com:5555" -key="worker1"

Then please add the following lines to your `$HOME/.ssh/config`.

	Host proxy-*
	ProxyCommand proxy -s ip-of-hub:5555 -k $(echo %h | sed 's/^proxy-//')

You can connect to `worker1` by `ssh proxy-worker1` now.
