#! /bin/sh

set -e

cd /home/user

git config --global user.email "holy@admin.rocks"
git config --global user.name "Holy Admin"

git init --bare .remote-example-100
git init --bare .remote-example-101

git clone /home/user/.remote-example-100 example-100
(
    cd example-100

    cat > README.md <<EOF
# A cool story about trains

The story is located in \`trains.md\`.
EOF
    touch trains.md
    git add README.md trains.md
    git commit -m "Add title to README

The story will be about trains but it has no content yet."

    cat >> README.md <<EOF

I like trains! Trains are great...
EOF
    git add README.md
    git commit -m "Write beginning of story

I really like trains :D"

    git push

    git reset --hard master~1
)

git clone /home/user/.remote-example-101 example-101
(
    cd example-101

    cat > README.md <<EOF
# This project contains a book in ships

The story is located in \`the-ship.md\`.
EOF

    cat > the-ship.md <<EOF
# The story of the ship

Once upon a time there was a ship on the sea.

The End.
EOF
    git add README.md the-ship.md
    git commit -m "Add README and beginning of ship story.

The story is quite short for now. It will be extended later on."

    cat > the-ship.md <<EOF
# The story of the ship

Once upon a time there was a ship on the stormy sea.
The wind was strong.

The End.
EOF
    git add the-ship.md
    git commit -m "Extend story a bit

The ship is in a storm."

    git push

    git reset --hard master~1
)

git config --global --unset-all user.email
git config --global --unset-all user.name
