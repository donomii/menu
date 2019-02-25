rm vort-fuse
go build github.com/donomii/menu
cp menu Menu.app/Contents/MacOS/
hdiutil create -volname Menu -srcfolder Menu.app/ -fs HFS+ -ov -format UDSB  -size 30m Menu_dist.dmg
mv Menu_dist.dmg.sparsebundle Menu_dist.dmg
hdiutil attach Menu_dist.dmg -mountpoint  mountpoint
cd mountpoint
ln -s ~/Applications
cd ..
hdiutil detach mountpoint
rm Menu.dmg
hdiutil convert Menu_dist.dmg  -format UDZO -o Menu.dmg
rm -r Menu_dist.dmg
