tunnelhub
=========

Unsafe tcp tunnel tool for massive control.

SSH Config
==========

`ssh_config` example:

	Host koala:*
	ProxyCommand proxy -k $(echo %h | sed 's/^koala://')
