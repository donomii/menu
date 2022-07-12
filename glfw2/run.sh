#!/bin/sh
go build .
cd ../voice
python transcribe.py | perl -pe '$|++;s/ and / /;s/ a / /;s/ the / /;' | python intent.py | ../glfw2/glfw2 -debug
#python transcribe.py > input.txt  &
#tail -f input.txt |  perl -pe '$|++;s/ and / /;s/ a / /;s/ the / /;' | python intent.py | ../glfw2/glfw2

wait

