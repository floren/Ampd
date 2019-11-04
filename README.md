# Ampd: An MPD client for Acme.

To build:

go get github.com/floren/Ampd

This will drop the Ampd binary into your $GOPATH/bin. Assuming that's in your $PATH, you can simply execute `Ampd` from within Acme.

Note that this is my first Acme application, so take all coding practices with a grain of salt.

## Flags

* `-server <host:port>` sets the MPD server. Default is "localhost:6600".
* `-password <pw>` sets the MPD password. Note that any user on your machine will be able to see the MPD password in `ps`.

## Basic Controls

The first window which appears is the main control window. The body contains controls for pausing and playing, skipping forward and backward, toggling repeat/random on and off, and opening the current playlist. *Middle-click* (execute) these controls to use them:

	[Prev] [Pause] [Play] [Next]
	[Random] [Repeat] [Playlist]
	Playing: The Who - Put The Money Down ("Odds And Sods")
	Random: 1, Repeat: 0

Note that the basic playback controls are also available in the tagline of the window, so you can keep the window "minimized" (only the tag visible) most of the time to save space.

## Playlist Management

Mid-clicking [Playlist] will open the playlist editor. This will display all songs in the current playlist. Each song is prefixed with its unique ID number. You can delete songs from the playlist by deleting them from the body of this window. Middle-click (execute) "WriteBack" in the window's tag line to save your changes to the current playlist. Executing "Reload" will refresh the playlist from the MPD server. Executing "Clear" will clear the playlist on the server. Adding a playlist name after "SavePlaylist" and executing e.g. "SavePlaylist foo" will save the current playlist as a playlist named "foo"; note that you must execute WriteBack before you save the playlist.

## Searching

You can search in two ways. First, right-clicking any text in the main playback control window will search for artists, albums, and song titles matching that text. If the current status is `Playing: Queen - Don't Stop Me Now ("Greatest Hits I")`, you can select "Queen" with the right mouse button and release to open a new search window containing any songs, albums, or artists matching "Queen".

You can search more specifically by specifying what type you want to query. For instance, to search only for the *album* named "Warren Zevon", type `album Warren Zevon` in the tag line and right-select it. You can search for "album", "artist", or "title" in this fashion.

Each kind of search brings up a new window showing search results; an example is shown below:

	0/ Warren Zevon - Frank And Jesse James [Warren Zevon]
	1/ Warren Zevon - Mama Couldn't Be Persuaded [Warren Zevon]
	2/ Warren Zevon - Backs Turned Looking Down The Path [Warren Zevon]
	3/ Warren Zevon - Hasten Down The Wind [Warren Zevon]
	4/ Warren Zevon - Poor, Poor Pitiful Me [Warren Zevon]
	5/ Warren Zevon - French Inhaler [Warren Zevon]
	6/ Warren Zevon - Mohammed's Radio [Warren Zevon]
	7/ Warren Zevon - I'll Sleep When I'm Dead [Warren Zevon]
	8/ Warren Zevon - Join Me In L.A. [Warren Zevon]
	9/ Warren Zevon - Desperados Under The Eaves [Warren Zevon]

Right-clicking the number preceding a song will add that song to the current playlist. Selecting multiple songs with the right button will add them *all* to the current playlist.

## Tricks

To add all songs, do a search on the string `artist` with no argument. This will pull back songs by every single artist. You can then select the entire search results buffer and right-click it to add all songs.
