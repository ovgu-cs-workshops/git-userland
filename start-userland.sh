#! /bin/sh

if [ ! -f "/home/user/.initializedHome" ]; then
    echo "Initializing Home"
    chown -R 1000:1000 /home/user
    su user -c "cp -r /home/user-template/. /home/user"
    chown -R 1000:1000 /home/user
fi

echo "Starting Userland"
su user -c "/bin/git-userland"
