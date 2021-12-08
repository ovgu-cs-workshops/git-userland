#! /bin/sh

set -e

cd /home/user

touch .initializedHome

git config --global user.email "holy@admin.rocks"
git config --global user.name "Holy Admin"

git init --bare .remote
git init --bare .remote-example-100
git init --bare .remote-example-101
git init --bare .remote-example-102
git init --bare .remote-example-201
git init --bare .remote-example-300


git clone /home/user/.remote-example-100 example-100-tmp

(
    cd example-100-tmp

    cat > README.md <<EOF
# A cool story about trains

> TODO
EOF
    git add README.md
    git commit -m "Add title to README

The story will be about trains but it has no content yet."

    git push
    git clone /home/user/.remote-example-100 ~/example-100

    sed -i 's/> TODO/I like trains! Trains are great. I once was on a train. It was great!\n\nThe End./g' README.md
    git add README.md
    git commit -m "Write beginning of story

I really like trains :D"

    git push
)
rm -r example-100-tmp

git clone /home/user/.remote-example-101 example-101-tmp
(
    cd example-101-tmp

    cat > README.md <<EOF
# A whole new story

Let's write about water bottles! The story is located in \`water-bottles.md\`.
EOF
    git add README.md
    git commit -m "Add README.md with project description

The readme tells the user where the story is located."

    git push
    git clone /home/user/.remote-example-101 ~/example-101

    cat >> README.md <<EOF

> For other stories please add new files.
EOF
    git add README.md
    git commit -m "Add note to README.md

Users should add new files for new stories."
    git push
)
rm -r example-101-tmp
(
    cd example-101

    cat > water-bottles.md <<EOF
# The story of the water bottle

Water is great. But when there is no water it is not great. There was once a water bottle without water. Someone filled it with water. The water bottle was happy again.

The end.
EOF
)

git clone /home/user/.remote-example-102 example-102-tmp
(
    cd example-102-tmp

    cat > README.md <<EOF
# This project contains a book on ships

The story is located in \`the-ship.md\`.
EOF

    cat > the-ship.md <<EOF
# The story of the ship

Once upon a time there was a ship on the sea.

The end.
EOF
    git add README.md the-ship.md
    git commit -m "Add README and beginning of ship story.

The story is quite short for now. It will be extended later on."

    git push
    git clone /home/user/.remote-example-102 ~/example-102

    cat > the-ship.md <<EOF
# The story of the ship

Once upon a time there was a ship on the stormy sea.
The wind was strong.

The end.
EOF
    git add the-ship.md
    git commit -m "Extend story a bit

The ship is in a storm."

    git push
)
rm -r example-102-tmp
(
    cd example-102

    cat > the-ship.md <<EOF
# The story of the ship

Once upon a time there was a blue ship on the sea.

The end.
EOF
)

git clone /home/user/.remote-example-300 example-300-tmp
(
    cd example-300-tmp

    cat > README.md <<EOF
# A whole new story

Let's write about water bottles! The story is located in \`water-bottles.md\`.
EOF
    git add README.md
    git commit -m "Add README.md with project description

The readme tells the user where the story is located."

    git push
    git clone /home/user/.remote-example-300 ~/example-300

    cat >> README.md <<EOF

> For other stories please add new files.
EOF
    git add README.md
    git commit -m "Add note to README.md

Users should add new files for new stories."
    git push
)
rm -r example-300-tmp
(
    cd example-300

    cat > water-bottles.md <<EOF
# The story of the water bottle

Water is great. But when there is no water it is not great. There was once a water bottle without water. Someone filled it with water. The water bottle was happy again.

The end.
EOF
)

git clone /home/user/.remote-example-201 example-201-tmp
(
    cd example-201-tmp

    cat > README.md <<EOF
# This project contains a sequel to the ship

The story is located in \`the-ship-2.md\`.
EOF

    git add README.md
    git commit -m "Initial commit"
    git push origin master

    git branch development
    git checkout development

    cat > the-ship-2.md <<EOF
# The Ship 2 - The Next Generation

In a not so far future there is a ship in space.
The ship is quite fast.
Now it is over there.

__fin__
EOF

    git add the-ship-2.md
    git commit -m "Add introduction

This should really be the content of a TV series. Might do that later"

    git push --set-upstream origin development
    git branch feature/second-chapter
    git checkout feature/second-chapter
    git push --set-upstream origin feature/second-chapter

    git checkout development
    git branch fix/ending
    git checkout fix/ending

    sed -i 's/__fin__/__The End__/g' the-ship-2.md
    git add -A
    git commit -m "Fix ending sentence

Our publisher said we cannot use french."
    git push --set-upstream origin fix/ending

    git checkout development
    git merge fix/ending
    git push --set-upstream origin development

    cd ..
)

rm -r example-201-tmp
git clone /home/user/.remote-example-201 example-201

git config --global --unset-all user.email
git config --global --unset-all user.name

# Set nano as default git editor with automatic linewrapping
git config --global core.editor "nano -r 70 -b"
