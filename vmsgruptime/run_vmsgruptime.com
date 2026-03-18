$ set noon
$! I have this set to run as my user, you probably want to change it
$! could be advisable to make a user for vmsgruptime
$ run/detach/process_name=vmsgruptime/uic=[users,mveety] vmsgruptime.exe
$ exit
