# askpw

interactive prompt wrapper for password managers

# synopsis

askpw \[--stderr\] \[--bin=PATH\] \[--entry=NAME\] \[OPTION\]...
askpw --version
askpw --help

# description

_askpw_ is intended to be used as *SSH_ASKPASS* command. instead of reading
the pass-phrase from the commandline input, a password manager is queried.
the database entry key is either read from the prompt or from the command-line
arguments. the remainder of the command-line is passed on to the password
manager command.

if no binary is specified, [pwsafe] is used. other password managers can be 
used as well (e.g. [passwordstore]), but have not been tested.

due to the forced appending of a non-flag argument, _askpw_ is not intended
to be used as a complete wrapper (e.g. `askpw --createdb` will not work)

[pwsafe]: http://nsd.dyndns.org/pwsafe/ "password database"
[passwordstore]: http://www.passwordstore.org "the standard unix password manager"

# arguments

all arguments which are not handled by _askpw_ as passed on to the sub-command
without modifications.

    --bin=PATH              the absolute path to the invoked binary
    --entry=NAME            this will not ask for the entry via prompt
    --stderr                ask for the entry key on stderr
    --version               display the askpw version and exit
    --help                  display the usage/help message and exit

using `--` as argument causes the remainder of the arguments to be passed on to
the sub-command, even if they are valid _askpw_ arguments.

# environment

the name of the entry to select can also be specified via the environment
variable **ASKPW_ENTRY**. it is only used, if no entry is specified on the
command-line.

# exit codes

an exit code of **1** indicates an error from the password manager command. all
other exit codes **>0** are used for internal errors by _askpw_ itself. zero is
returned upon successful execution.

# examples

    #!/bin/sh
    export SSH_ASKPASS="askpw --bin=/usr/local/bin/passtore --entry=$USER"
    export DISPLAY=null:0
    setsid ssh "$@"

this wrapper script for ssh receives the key passphrase from the _pwsafe_
database. it is assumed, that _passtore_ does not require any interaction,
as there will be no terminal available to receive input.

> askpw --stderr --echo | decrypt --stdin

prompt for the password entry on *stderr*. _pwsafe_ will emit the password
on *stdout* for the receiving command to read.

> askpw --entry=alias --add

this would be the same as running _pwsafe_ directly

> askpw --bin=/bin/pass -- --version

same as running `/bin/pass --version`

## askpass

the accompaning _askpass_ script can be used in combination with _askpw_.
symlinking or aliasing it allows password based access (or via passphrase
protected keyfile) in conjunction with a password manager.

internally is defines the **SSH_ASKPASS** environment variable with a
self-referencing version of _askpass_ which holds the actual password.

_askpass_ is just a wrapper script. the actual command is resolved
in two ways:

**ASKPASS_CONSUMER**  
if defined, its values is executed together with the arguments passed
initially to _askpass_.  

> $> ASKPASS_CONSUMER=scp askpass backup.bin remote:/backups

**$0**  
the name of the wrapper script is stripped of the following substrings:  
ask, askpw, askpass  
the remainder is the wrapped command name.

> $> echo "export ASKPASS_COMMAND="askpw --stderr"" >> $HOME/.profile  
> $> echo "export PATH=\$HOME/bin:\$PATH" >> $HOME/.profile  
> $> mkdir $HOME/bin  
> $> ln -s /usr/local/bin/askpass $HOME/bin/ssh-ask  
> $> ln -s /usr/local/bin/askpass $HOME/bin/scp-ask  
> $> ln -s /usr/local/bin/askpass $HOME/bin/sftp-ask  

log out and in again (or source _.profile_) and you are able to ssh into 
systems with passphrase protected keys: `ssh-ask -T git@github.com`

--

the actual password is provided by an external source

**ASKPASS_ARGUMENTS**  
used _askpw_ with the defined arguments to extract the password

**ASKPASS_COMMAND**  
executes the value of the variable as command.  
the password is read from stdout

**stdin**  
if none of the previous variables exist, the password is read from the
input and passed on to the wrapped command.
