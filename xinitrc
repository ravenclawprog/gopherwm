# xinitrc file for testing our window manager.
./bin/gopherwm &
# 1. Start some programs to play with.
xclock -geometry 100x100+450+0 &
xeyes &
sleep 2 &
xterm
