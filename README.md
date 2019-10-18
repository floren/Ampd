An MPD client for Acme.

To build:

go get github.com/floren/Ampd

This will drop the Ampd binary into your $GOPATH/bin. Assuming that's in your $PATH, you can simply execute `Ampd` from within Acme.

The first window which appears is the main control window. The body contains controls for pausing and playing, skipping forward and backward, toggling repeat/random on and off, and opening the current playlist. *Right-click* these controls to use them.

Right-clicking [Playlist] will open the playlist editor. This will display all songs in the current playlist. Each song is prefixed with its unique ID number. You can delete songs from the playlist by deleting them from the body of this window. Middle-click (execute) "WriteBack" in the window's tag line to save your changes. Executing "Reload" will refresh the playlist from the MPD server. Executing "Clear" will clear the playlist on the server. Adding a playlist name after "SavePlaylist" and executing e.g. "SavePlaylist foo" will save the current playlist as a playlist named "foo"; note that you must execute WriteBack before you save the playlist.